package models

import (
	"regexp"
	"time"
	"unicode"
	"test123/errors"
)

// User represents the core user model.
type User struct {
	ID           int       `json:"id"`  //auto generate no need
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Username     string    `json:"username"`
	Password     string    `json:"password,omitempty"` // omit in JSON responses
	MobileNumber string    `json:"mobile_number"`
	CreatedAt    time.Time `json:"created_at"`
}

// ===========================
//  VALIDATION
// ===========================
func (u *User) Validate() error {
	if u.Name == "" || u.Email == "" || u.Username == "" || u.Password == "" {
		return errors.ErrMissingField
	}

	//  Validate Email Format
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(u.Email) {
		return errors.ErrInvalidEmail
	}

	//  Validate Mobile Number (Optional but must be 10 digits if provided)
	if u.MobileNumber != "" {
		mobileRegex := regexp.MustCompile(`^[0-9]{10}$`)
		if !mobileRegex.MatchString(u.MobileNumber) {
			return errors.ErrInvalidField
		}
	}

	//  Validate Password Strength (at least one upper, lower, digit, symbol, min 8)
	if err := validatePassword(u.Password); err != nil {
		return err
	}

	return nil
}

// ===========================
// PASSWORD STRENGTH CHECKER
// ===========================
func validatePassword(password string) error {
	var (
		hasMinLen  = len(password) >= 8
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasNumber = true
		case unicode.IsPunct(ch) || unicode.IsSymbol(ch):
			hasSpecial = true
		}
	}

	if !(hasMinLen && hasUpper && hasLower && hasNumber && hasSpecial) {
		return errors.ErrWeakPassword
	}
	return nil
}
