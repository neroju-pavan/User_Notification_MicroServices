package jwt

import (
	"fmt"
	"time"

	"log"

	"github.com/golang-jwt/jwt/v5"
)

type Jwt struct {
	SecretKey     string
	SecretKeyByte []byte
}

func NewJwt(secret string) *Jwt {
	return &Jwt{
		SecretKey:     secret,
		SecretKeyByte: []byte(secret),
	}
}



func (J *Jwt) GenerateJWTtoken(Id int, duration int) (string, error) {

	//claims creation and adding to token

	claims := jwt.MapClaims{
		"user": Id,
		"exp":  time.Now().Add(time.Duration(duration) * time.Minute).Unix(),

		"iat": time.Now().Unix(),
		
	}
	//New With Claims give *JWTToken (its aint string)
	token := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)

	tokenString, err := token.SignedString([]byte(J.SecretKey))
	if err != nil {
		return "", err
	}
	log.Println(token.Header)
	log.Println(token.Claims)
	log.Println(token.Signature)
	return tokenString, nil

}
func (j *Jwt) Decode(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.SecretKeyByte, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		return nil, jwt.ErrTokenInvalidClaims
	}

	if !token.Valid {
		return nil, jwt.ErrTokenInvalidId
	}

	return claims, nil
}

func (j *Jwt) FetchClaim(key string, claims jwt.MapClaims) string {
	if value, exists := claims[key].(string); exists && value != "" {
		return value
	}
	return ""
}

func (j *Jwt) GetExpiryFromToken(tokenStr string) (int64, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenStr, jwt.MapClaims{})
	if err != nil {
		return 0, err
	}

	claims := token.Claims.(jwt.MapClaims)
	exp := int64(claims["exp"].(float64))
	return exp, nil
}
