package models

import (
	"test123/errors"
	"time"
)

// UserProfile represents extended profile information for a user.
// It references the core User by UserID.
type UserProfile struct {
	ID          int               `json:"id"`
	UserID      int               `json:"user_id"`
	Bio         string            `json:"bio,omitempty"`
	AvatarURL   string            `json:"avatar_url,omitempty"`
	Location    string            `json:"location,omitempty"`
	DOB         *time.Time        `json:"dob,omitempty"`
	Preferences map[string]string `json:"preferences,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// Validate checks required fields for UserProfile.
func (p *UserProfile) Validate() error {
	if p.UserID == 0 {
		return errors.ErrMissingField
	}
	return nil
}
