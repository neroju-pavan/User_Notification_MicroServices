package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	kafka "test123/kafka/producers"
	"test123/logger"
	"test123/utils"
	"test123/utils/jwt"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type AuthService struct {
	UserService *UserService
	Redis       *redis.Client
	JWT         *jwt.Jwt
	prod        *kafka.KafkaNotificationProducer
}

func NewAuthService(userService *UserService, redisClient *redis.Client, jwt *jwt.Jwt, producer *kafka.KafkaNotificationProducer) *AuthService {
	return &AuthService{
		UserService: userService,
		Redis:       redisClient,
		JWT:         jwt,
		prod:        producer,
	}
}

// GenerateResetToken generates a password reset token and sends it via email.
func (s *AuthService) GenerateResetToken(ctx context.Context, email string) (int, map[string]string) {
	logger.Info("GenerateResetToken", "called", map[string]interface{}{"email": email})

	if email == "" {
		logger.Error("GenerateResetToken", "email required", nil)
		return 400, map[string]string{"error": "email required"}
	}

	user, err := s.UserService.GetByUserByEmail(ctx, email)
	if err != nil {
		logger.Error("GenerateResetToken", "user not found", map[string]interface{}{"email": email})
		return 404, map[string]string{"error": "user not found"}
	}

	if existingToken, _ := s.Redis.Get(ctx, "reset:active:"+user.Username).Result(); existingToken != "" {
		logger.Info("GenerateResetToken", "active token exists", map[string]interface{}{"username": user.Username})
		return 400, map[string]string{"error": "A reset link is already sent. Please check your email."}
	}

	// Generate token
	tokenBytes := make([]byte, 32)
	rand.Read(tokenBytes)
	token := base64.URLEncoding.EncodeToString(tokenBytes)
	resetURL := fmt.Sprintf("http://localhost:8083/api/v1/auth/reset-password?token=%s", token)

	event := utils.NewEmailNotificationEvent(
		user.ID,
		"Token",
		"Password Reset",
		"Click the link to reset your password. Token expires in 10 minutes.",
		user.Email,
		map[string]string{"token": resetURL},
	)
	event.EventID = uuid.New()
	event.CreatedAt = time.Now()
	event.NotificationType = "email"

	eventBytes, err := json.Marshal(event)
	if err != nil {
		logger.Error("GenerateResetToken", "failed to serialize event", map[string]interface{}{"error": err.Error()})
		return 500, map[string]string{"error": "failed to serialize event"}
	}

	if err := s.Redis.Set(ctx, "reset_token:"+token, user.Username, 10*time.Minute).Err(); err != nil {
		logger.Error("GenerateResetToken", "failed to store reset_token in redis", map[string]interface{}{"error": err.Error()})
		return 500, map[string]string{"error": "internal server error"}
	}

	if err := s.Redis.Set(ctx, "reset:active:"+user.Username, token, 10*time.Minute).Err(); err != nil {
		logger.Error("GenerateResetToken", "failed to store active token in redis", map[string]interface{}{"error": err.Error()})
		return 500, map[string]string{"error": "failed to store active token reference"}
	}

	s.Redis.Set(ctx, "reset:invalid:"+user.Username, 0, 10*time.Minute)

	if err := s.prod.Send(ctx, eventBytes); err != nil {
		logger.Error("GenerateResetToken", "failed to send event to Kafka", map[string]interface{}{"error": err.Error()})
		return 500, map[string]string{"error": "failed to send notification event"}
	}

	logger.Info("GenerateResetToken", "reset link sent successfully", map[string]interface{}{"username": user.Username, "token": token})
	return 200, map[string]string{"message": "reset link sent successfully"}
}

// ResetPassword validates the token and updates the user's password.
func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) (int, map[string]string) {
	logger.Info("ResetPassword", "called", map[string]interface{}{"token": token})

	if token == "" {
		return 400, map[string]string{"error": "token required"}
	}

	username, err := s.Redis.Get(ctx, "reset_token:"+token).Result()
	if err != nil {
		logger.Error("ResetPassword", "invalid or expired token", map[string]interface{}{"token": token})
		return 400, map[string]string{"error": "invalid or expired token"}
	}

	key := "reset:invalid:" + username
	count, _ := s.Redis.Get(ctx, key).Int()
	if count >= 1 {
		logger.Error("ResetPassword", "too many invalid attempts", map[string]interface{}{"username": username})
		return 429, map[string]string{"error": "Too many invalid attempts. Request a new reset link."}
	}

	if newPassword == "" {
		s.Redis.Incr(ctx, key)
		logger.Error("ResetPassword", "password required", map[string]interface{}{"username": username})
		return 400, map[string]string{"error": "password required"}
	}

	if err := s.UserService.UpdatePassword(ctx, username, newPassword); err != nil {
		s.Redis.Incr(ctx, key)
		logger.Error("ResetPassword", "failed to update password", map[string]interface{}{"username": username, "error": err.Error()})
		return 500, map[string]string{"error": "failed to update password"}
	}

	// Cleanup Redis keys
	s.Redis.Del(ctx, "reset_token:"+token)
	s.Redis.Del(ctx, "reset:active:"+username)
	s.Redis.Del(ctx, key)
	s.Redis.Del(ctx, "access_token:"+username)

	logger.Info("ResetPassword", "password reset successful", map[string]interface{}{"username": username})
	return 200, map[string]string{"message": "password reset successful"}
}

// Login authenticates a user and returns JWT tokens.
func (s *AuthService) Login(ctx context.Context, username, password string) (int, map[string]interface{}) {
	logger.Info("Login", "called", map[string]interface{}{"username": username})

	if username == "" || password == "" {
		logger.Error("Login", "username or password missing", nil)
		return 400, map[string]interface{}{"error": "username and password required"}
	}

	user, err := s.UserService.GetUserByEmailOrUsername(ctx, username)
	if err != nil {
		logger.Error("Login", "user not found", map[string]interface{}{"username": username})
		return 404, map[string]interface{}{"error": "user not found"}
	}

	count, _ := s.Redis.Get(ctx, "attempt_key:"+username).Int()
	if count > 5 {
		s.Redis.Expire(ctx, "attempt_key:"+username, 10*time.Minute)

		event := utils.NewEmailNotificationEvent(
			user.ID,
			"security",
			"Verify is it you",
			"Someone tried to login to your account",
			user.Email,
			nil,
		)
		eventBytes, _ := json.Marshal(event)
		s.prod.Send(ctx, eventBytes)

		return 429, map[string]interface{}{"error": "too many requests, try after 10 minutes"}
	}

	if user.Password != password {
		s.Redis.Incr(ctx, "attempt_key:"+username)
		logger.Error("Login", "wrong password", map[string]interface{}{"username": username})
		return 401, map[string]interface{}{"error": "wrong password"}
	}

	access, err := s.JWT.GenerateJWTtoken(user.ID, 15)
	if err != nil {
		logger.Error("Login", "failed to generate access token", map[string]interface{}{"username": username, "error": err.Error()})
		return 500, map[string]interface{}{"error": "failed to generate access token"}
	}

	refresh, err := s.JWT.GenerateJWTtoken(user.ID, 240)
	if err != nil {
		logger.Error("Login", "failed to generate refresh token", map[string]interface{}{"username": username, "error": err.Error()})
		return 500, map[string]interface{}{"error": "failed to generate refresh token"}
	}

	s.Redis.Set(ctx, "access_token:"+user.Username, access, 15*time.Minute)
	s.Redis.Set(ctx, "refresh_token:"+user.Username, refresh, 4*time.Hour)

	logger.Info("Login", "login successful", map[string]interface{}{"username": username})
	return 200, map[string]interface{}{"access_token": access, "refresh_token": refresh}
}

// WipeOutSession logs out the user by deleting tokens from Redis.
func (s *AuthService) WipeOutSession(ctx context.Context, accessToken, refreshToken string) (int, map[string]string) {
	logger.Info("WipeOutSession", "called", nil)

	if accessToken == "" || refreshToken == "" {
		logger.Error("WipeOutSession", "access token or refresh token missing", nil)
		return 400, map[string]string{"error": "access token and refresh token required"}
	}

	claims, err := s.JWT.Decode(accessToken)
	if err != nil {
		logger.Error("WipeOutSession", "invalid access token", map[string]interface{}{"error": err.Error()})
		return 401, map[string]string{"error": "invalid access token"}
	}

	userID := int(claims["user"].(float64))
	user, err := s.UserService.GetUserByID(ctx, userID)
	if err != nil {
		logger.Error("WipeOutSession", "user not found", map[string]interface{}{"userID": userID})
		return 404, map[string]string{"error": "user not found"}
	}

	storedToken, err := s.Redis.Get(ctx, "access_token:"+user.Username).Result()
	if err != nil || storedToken != accessToken {
		return 400, map[string]string{"error": "token already expired or logged out"}
	}
	s.Redis.Del(ctx, "access_token:"+user.Username)

	if _, err := s.JWT.Decode(refreshToken); err != nil {
		return 400, map[string]string{"error": "invalid or expired refresh token"}
	}

	storedRefresh, err := s.Redis.Get(ctx, "refresh_token:"+user.Username).Result()
	if err != nil || storedRefresh != refreshToken {
		return 403, map[string]string{"error": "refresh token invalid or expired"}
	}
	s.Redis.Del(ctx, "refresh_token:"+user.Username)

	logger.Info("WipeOutSession", "logout successful", map[string]interface{}{"username": user.Username})
	return 200, map[string]string{"message": "user logged out successfully"}
}

// GenerateAccessToken generates a new access token using a valid refresh token.
func (s *AuthService) GenerateAccessToken(ctx context.Context, refreshToken string) (int, map[string]string) {
	logger.Info("GenerateAccessToken", "called", nil)

	if refreshToken == "" {
		return 400, map[string]string{"error": "refresh token required"}
	}

	claims, err := s.JWT.Decode(refreshToken)
	if err != nil {
		return 401, map[string]string{"error": "invalid refresh token"}
	}

	userID := int(claims["user"].(float64))
	user, err := s.UserService.GetUserByID(ctx, userID)
	if err != nil {
		return 404, map[string]string{"error": "user not found"}
	}

	storedRefresh, err := s.Redis.Get(ctx, "refresh_token:"+user.Username).Result()
	if err != nil || storedRefresh != refreshToken {
		return 403, map[string]string{"error": "refresh token invalid or expired"}
	}

	newAccess, err := s.JWT.GenerateJWTtoken(userID, 15)
	if err != nil {
		return 500, map[string]string{"error": "failed to generate access token"}
	}

	s.Redis.Set(ctx, "access_token:"+user.Username, newAccess, 15*time.Minute)
	logger.Info("GenerateAccessToken", "new access token generated", map[string]interface{}{"username": user.Username})
	return 200, map[string]string{"access_token": newAccess}
}

// Authorize validates the access token in the HTTP request.
func (s *AuthService) Authorize(ctx context.Context, r *http.Request) (string, error) {
	logger.Info("Authorize", "called", nil)

	token, err := utils.ExtractToken(r)
	if err != nil || token == "" {
		return "", errors.New("missing token")
	}

	claims, err := s.JWT.Decode(token)
	if err != nil {
		return "", errors.New("invalid token")
	}

	userID := int(claims["user"].(float64))
	user, err := s.UserService.GetUserByID(ctx, userID)
	if err != nil {
		return "", errors.New("user not found")
	}

	storedToken, err := s.Redis.Get(ctx, "access_token:"+user.Username).Result()
	if err != nil || storedToken != token {
		return "", errors.New("token expired or logged out")
	}

	logger.Info("Authorize", "authorization successful", map[string]interface{}{"username": user.Username})
	return strconv.Itoa(userID), nil
}
