package errors

import "errors"

// 400 – Bad Request
var (
	ErrInvalidJSON        = errors.New("invalid JSON format")
	ErrMissingField       = errors.New("required field missing")
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrInvalidPassword    = errors.New("invalid password format")
	ErrWeakPassword       = errors.New("weak password")
	ErrInvalidField       = errors.New("invalid field")
	ErrInvalidCategory    = errors.New("invalid category data")
	ErrInvalidParams      = errors.New("invalid parameters")
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrInvalidToken       = errors.New("invalid token")
	ErrBadRequest         = errors.New("bad request")
	ErrValidationFailed   = errors.New("validation failed")
)

// 401 – Unauthorized
var (
	ErrUnauthorized      = errors.New("unauthorized")
	ErrMissingAuthHeader = errors.New("missing authorization header")
	ErrInvalidAuthHeader = errors.New("invalid authorization header format")
	ErrInvalidJWT        = errors.New("invalid JWT token")
	ErrExpiredJWT        = errors.New("expired JWT token")
)

// 403 – Forbidden
var (
	ErrForbidden      = errors.New("forbidden")
	ErrRoleNotAllowed = errors.New("role not allowed for this action")
	ErrAccessDenied   = errors.New("access denied")
)

// 404 – Not Found
var (
	ErrUserNotFound     = errors.New("user not found")
	ErrCategoryNotFound = errors.New("category not found")
	ErrResourceNotFound = errors.New("resource not found")
)

// 409 – Conflict
var (
	ErrUserExists       = errors.New("user already exists")
	ErrCategoryExists   = errors.New("category already exists")
	ErrAlreadyProcessed = errors.New("request already processed")
	ErrDuplicateRequest = errors.New("duplicate request")
)

// 400 – Bad Request
var (
	ErrMissingCredentials     = errors.New("missing username or password")
	ErrInvalidPasswordAttempt = errors.New("wrong password")
)

// 429 – Too Many Requests
var (
	ErrTooManyLoginAttempts = errors.New("too many login attempts")
	ErrTooManyResetAttempts = errors.New("too many reset attempts")
)

// 422 – Unprocessable Entity
// (Already included above)

// 429 – Too Many Requests
var ErrRateLimitExceeded = errors.New("rate limit exceeded")

// 500 – Internal Errors
var (
	ErrDatabaseFailure = errors.New("database operation failed")
	ErrCacheFailure    = errors.New("cache operation failed")
	ErrInternalFailure = errors.New("internal server error")
	ErrUnknown         = errors.New("unknown server error")
)

// 503 – Service Unavailable
var (
	ErrServiceUnavailable = errors.New("service temporarily unavailable")
	ErrTimeout            = errors.New("operation timed out")
	ErrDependencyFailure  = errors.New("dependency service failed")
)
