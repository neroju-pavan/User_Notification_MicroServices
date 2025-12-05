package service

import (
	"context"

	"test123/errors"
	"test123/logger"
	"test123/models"
	"test123/repositories"
)

type ProfileService struct {
	Repo repositories.ProfileRepoInterface
}

func NewProfileService(repo repositories.ProfileRepoInterface) *ProfileService {
	return &ProfileService{Repo: repo}
}

func (s *ProfileService) CreateProfile(ctx context.Context, p models.UserProfile) error {
	logger.Info("ProfileService.CreateProfile", "service create profile", map[string]interface{}{"user_id": p.UserID})

	if p.UserID == 0 {
		logger.Warn("ProfileService.CreateProfile", "missing user id")
		return errors.ErrMissingField
	}

	return s.Repo.CreateProfile(ctx, p)
}

func (s *ProfileService) GetProfileByUserID(ctx context.Context, userID int) (*models.UserProfile, error) {
	logger.Info("ProfileService.GetProfileByUserID", "service get profile", map[string]interface{}{"user_id": userID})

	if userID == 0 {
		return nil, errors.ErrMissingField
	}

	return s.Repo.GetProfileByUserID(ctx, userID)
}

func (s *ProfileService) UpdateProfile(ctx context.Context, p models.UserProfile) error {
	if p.UserID == 0 {
		return errors.ErrMissingField
	}
	return s.Repo.UpdateProfile(ctx, p)
}
