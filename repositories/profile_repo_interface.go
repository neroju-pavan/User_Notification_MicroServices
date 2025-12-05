package repositories

import (
	"context"
	"test123/models"
)

type ProfileRepoInterface interface {
	CreateProfile(ctx context.Context, p models.UserProfile) error
	GetProfileByUserID(ctx context.Context, userID int) (*models.UserProfile, error)
	UpdateProfile(ctx context.Context, p models.UserProfile) error
}
