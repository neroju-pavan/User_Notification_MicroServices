package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"test123/errors"
	kafka "test123/kafka/producers"
	"test123/logger"
	"test123/models"
	"test123/repositories"
	"test123/utils"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/willf/bloom"
)

type UserService struct {
	UserRepo     repositories.UserRepoInterface
	prod         *kafka.KafkaNotificationProducer
	UserRoleRepo repositories.UserRoleRepoInterface
	Redis        *redis.Client
	Bloom        *bloom.BloomFilter
}

// Constructor
func NewUserService(repo repositories.UserRepoInterface, Prod *kafka.KafkaNotificationProducer, userRoleRepo repositories.UserRoleRepoInterface, redis *redis.Client, bf *bloom.BloomFilter) *UserService {
	return &UserService{
		UserRepo:     repo,
		prod:         Prod,
		UserRoleRepo: userRoleRepo,
		Redis:        redis,
		Bloom:        bf,
	}
}

func (s *UserService) GetAllUsers(ctx context.Context) ([]models.User, error) {
	logger.Info("GetAllUsers", "Service call received")
	return s.UserRepo.GetAllUsers(ctx)
}

func (s *UserService) CreateUser(ctx context.Context, user models.User) error {

	logger.Info("CreateUser", "Creating user", map[string]interface{}{
		"email": user.Email,
	})

	// Validation
	if user.Name == "" {
		logger.Warn("CreateUser", "missing name")
		return fmt.Errorf("%w: name is required", errors.ErrMissingField)
	}
	if user.Email == "" {
		logger.Warn("CreateUser", "missing email")
		return fmt.Errorf("%w: email is required", errors.ErrMissingField)
	}

	// Save user
	err := s.UserRepo.CreateUser(ctx, user)
	if err != nil {
		logger.Error("CreateUser", "DB create failed", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	res, err := s.UserRepo.GetUserByEmail(ctx, user.Email)
	if err != nil {
		logger.Error("CreateUser", "Failed to fetch created user", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	// Build notification event
	event := utils.NewEmailNotificationEvent(
		res.ID,
		"user_created",
		"Successful Account Creation",
		"Hi "+res.Username+", your account has been created successfully.",
		res.Email,
		nil,
	)

	event.EventID = uuid.New()
	event.CreatedAt = time.Now()
	event.NotificationType = "email"

	message, err := json.Marshal(event)
	if err != nil {
		logger.Error("CreateUser", "Failed to marshal Kafka event", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to marshal notification event: %v", err)
	}

	// Publish to Kafka
	if err := s.prod.Send(ctx, message); err != nil {
		logger.Error("CreateUser", "Kafka publish failed", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to publish notification event: %v", err)
	}

	// Assign role
	if err := s.UserRoleRepo.AddUserRole(ctx, "user", res.ID); err != nil {
		logger.Error("CreateUser", "Failed to assign role", map[string]interface{}{
			"user_id": res.ID,
			"error":   err.Error(),
		})
		return fmt.Errorf("failed to add role: %v", err)
	}

	logger.Info("CreateUser", "User created successfully", map[string]interface{}{
		"user_id": res.ID,
	})

	return nil
}

func (s *UserService) Get(ctx context.Context) ([]models.User, error) {
	logger.Info("GetUsers", "Fetching all users")
	return s.UserRepo.GetAllUsers(ctx)
}

func (s *UserService) GetUserByID(ctx context.Context, id int) (*models.User, error) {

	logger.Info("GetUserByID", "Fetching user", map[string]interface{}{
		"id": id,
	})

	if id <= 0 {
		logger.Warn("GetUserByID", "Invalid user ID", map[string]interface{}{"id": id})
		return nil, fmt.Errorf("%w: invalid user ID", errors.ErrMissingField)
	}

	return s.UserRepo.GetUserByID(ctx, id)
}

func (s *UserService) UpdateUser(ctx context.Context, user models.User) error {

	logger.Info("UpdateUser", "Updating user", map[string]interface{}{
		"id": user.ID,
	})

	if user.ID == 0 {
		logger.Warn("UpdateUser", "missing ID")
		return fmt.Errorf("%w: missing user ID", errors.ErrMissingField)
	}
	if user.Name == "" && user.Email == "" {
		logger.Warn("UpdateUser", "No fields to update")
		return fmt.Errorf("%w: nothing to update", errors.ErrMissingField)
	}

	return s.UserRepo.UpdateUser(ctx, user)
}

func (s *UserService) DeleteUser(ctx context.Context, id int) error {

	logger.Info("DeleteUser", "Deleting user", map[string]interface{}{
		"id": id,
	})

	if id <= 0 {
		logger.Warn("DeleteUser", "invalid id", map[string]interface{}{
			"id": id,
		})
		return fmt.Errorf("%w: invalid user ID", errors.ErrMissingField)
	}

	return s.UserRepo.DeleteUser(ctx, id)
}

func (s *UserService) GetByUserByEmail(ctx context.Context, email string) (*models.User, error) {

	logger.Info("GetUserByEmail", "Fetching user", map[string]interface{}{
		"email": email,
	})

	return s.UserRepo.GetUserByEmail(ctx, email)
}

func (s *UserService) UpdatePassword(ctx context.Context, email string, password string) error {

	logger.Info("UpdatePassword", "Updating password", map[string]interface{}{
		"email": email,
	})

	return s.UserRepo.UpdatePassword(ctx, email, password)
}

func (s *UserService) GetUserByEmailOrUsername(ctx context.Context, key string) (*models.User, error) {

	logger.Info("GetUserByEmailOrUsername", "Fetching user", map[string]interface{}{
		"key": key,
	})

	return s.UserRepo.GetUserByEmailOrUsername(ctx, key)
}

func (s *UserService) UsernameExists(ctx context.Context, username string) (bool, error) {

	logger.Info("UsernameExists", "Checking username", map[string]interface{}{
		"username": username,
	})

	if username == "" {
		return false, nil
	}

	// BLOOM FILTER
	if !s.Bloom.TestString(username) {
		logger.Debug("UsernameExists", "Bloom filter indicates NOT exists")
		return false, nil
	}

	// REDIS
	_, err := s.Redis.Get(ctx, username).Result()
	if err == nil {
		logger.Info("UsernameExists", "Found in Redis cache")
		return true, nil
	}

	if err != nil && err != redis.Nil {
		logger.Warn("UsernameExists", "Redis error", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// DATABASE CHECK
	user, err := s.UserRepo.GetUserByUsername(ctx, username)
	if err != nil {
		if err == errors.ErrUserNotFound {
			logger.Info("UsernameExists", "Not found in DB")
			return false, nil
		}
		logger.Error("UsernameExists", "DB error", map[string]interface{}{
			"error": err.Error(),
		})
		return false, err
	}

	// caching again
	s.Bloom.AddString(user.Username)
	s.Redis.Set(ctx, user.Username, 1, 10*time.Minute)

	logger.Info("UsernameExists", "User exists", map[string]interface{}{
		"username": username,
	})

	return true, nil
}

func (s *UserService) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {

	logger.Info("GetUserByUsername", "Fetching user", map[string]interface{}{
		"username": username,
	})

	return s.UserRepo.GetUserByUsername(ctx, username)
}

func (s *UserService) GetUsersWithFiltersCursor(
	ctx context.Context,
	limit int,
	cursorStr, search, fromStr, toStr string,
) ([]models.User, *time.Time, error) {

	logger.Info("GetUsersWithFiltersCursor", "Fetching filtered users")

	var cursor *time.Time
	cursorStr = strings.TrimSpace(cursorStr)
	if cursorStr != "" {
		t, err := time.Parse(time.RFC3339, cursorStr)
		if err != nil {
			logger.Warn("GetUsersWithFiltersCursor", "Invalid cursor format", map[string]interface{}{
				"value": cursorStr,
			})
			return nil, nil, fmt.Errorf("%w: invalid cursor format", errors.ErrInvalidCategory)
		}
		cursor = &t
	}

	var fromDate, toDate *time.Time

	if fromStr != "" {
		t, err := time.Parse("2006-01-02", fromStr)
		if err != nil {
			logger.Warn("GetUsersWithFiltersCursor", "Invalid from date", map[string]interface{}{
				"value": fromStr,
			})
			return nil, nil, fmt.Errorf("%w: invalid from date", errors.ErrInvalidCategory)
		}
		fromDate = &t
	}

	if toStr != "" {
		t, err := time.Parse("2006-01-02", toStr)
		if err != nil {
			logger.Warn("GetUsersWithFiltersCursor", "Invalid to date", map[string]interface{}{
				"value": toStr,
			})
			return nil, nil, fmt.Errorf("%w: invalid to date", errors.ErrInvalidCategory)
		}
		toDate = &t
	}

	if fromDate != nil && toDate != nil && fromDate.After(*toDate) {
		logger.Warn("GetUsersWithFiltersCursor", "fromDate > toDate")
		return nil, nil, fmt.Errorf("%w: from date cannot be after to date", errors.ErrInvalidCategory)
	}

	if limit <= 0 || limit > 100 {
		limit = 5
	}

	return s.UserRepo.GetUsersWithFiltersCursor(ctx, limit, cursor, search, fromDate, toDate)
}
