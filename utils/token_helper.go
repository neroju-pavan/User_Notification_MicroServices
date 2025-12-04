package utils

import (
	"fmt"
	"net/http"
	"strings"
)

type TokenReq struct {
	Token string `json:"token"`
}

func ExtractToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")

	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer "), nil
	}

	return "", fmt.Errorf("missing token in Authorization header")
}
