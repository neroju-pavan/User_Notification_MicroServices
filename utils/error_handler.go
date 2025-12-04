package utils

import (
	"encoding/json"

	"net/http"
	e "test123/errors"
)

// JSON response helper (EXPORTED)
func RespondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func HttpStatusFromError(err error) int {
	switch err {

	// 400
	case e.ErrInvalidJSON, e.ErrMissingField, e.ErrInvalidEmail, e.ErrWeakPassword,
		e.ErrInvalidField, e.ErrInvalidCategory, e.ErrInvalidParams, e.ErrInvalidCredentials,
		e.ErrInvalidToken, e.ErrBadRequest, e.ErrValidationFailed:
		return http.StatusBadRequest

	// 401
	case e.ErrUnauthorized, e.ErrMissingAuthHeader, e.ErrInvalidAuthHeader,
		e.ErrInvalidJWT, e.ErrExpiredJWT:
		return http.StatusUnauthorized

	// 403
	case e.ErrForbidden, e.ErrRoleNotAllowed, e.ErrAccessDenied:
		return http.StatusForbidden

	// 404
	case e.ErrUserNotFound, e.ErrCategoryNotFound, e.ErrResourceNotFound:
		return http.StatusNotFound

	// 409
	case e.ErrUserExists, e.ErrCategoryExists, e.ErrAlreadyProcessed, e.ErrDuplicateRequest:
		return http.StatusConflict

	// 422
	case e.ErrValidationFailed:
		return http.StatusUnprocessableEntity

	// 429
	case e.ErrRateLimitExceeded:
		return http.StatusTooManyRequests

	case e.ErrMissingCredentials, e.ErrInvalidPasswordAttempt:
		return http.StatusBadRequest

	case e.ErrTooManyLoginAttempts, e.ErrTooManyResetAttempts:
		return http.StatusTooManyRequests

	// 503
	case e.ErrServiceUnavailable, e.ErrTimeout, e.ErrDependencyFailure:
		return http.StatusServiceUnavailable

	// 500 default
	case e.ErrDatabaseFailure, e.ErrCacheFailure, e.ErrInternalFailure, e.ErrUnknown:
		return http.StatusInternalServerError

	default:
		return http.StatusInternalServerError
	}
}

func HttpStatusFromSuccess(action string) int {
	switch action {

	case "created", "register", "signup", "add":
		return http.StatusCreated // 201

	case "accepted", "queued", "processing":
		return http.StatusAccepted // 202

	case "deleted", "removed":
		return http.StatusNoContent // 204

	default:
		return http.StatusOK // 200
	}
}
