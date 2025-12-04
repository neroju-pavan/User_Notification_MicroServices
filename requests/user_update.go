package requests

import (
	"regexp"
	"test123/errors"
)

type UserReq struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	Username     string `json:"username"`
	MobileNumber string `json:"mobile_number"`
}

func (u *UserReq) Validate() error {
	if u.Name == "" || u.Email == "" || u.Username == "" {
		return errors.ErrMissingField
	}

	// Email format check
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(u.Email) {
		return errors.ErrInvalidEmail
	}

	// Mobile must be 10 digits (optional but must be valid)
	if u.MobileNumber != "" {
		mobileRegex := regexp.MustCompile(`^[0-9]{10}$`)
		if !mobileRegex.MatchString(u.MobileNumber) {
			return errors.ErrInvalidField
		}
	}

	return nil
}
