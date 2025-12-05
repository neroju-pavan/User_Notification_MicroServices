package repositories

import (
	"context"
	"fmt"
	"time"

	"test123/errors"
	"test123/logger"
	"test123/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProfileRepo struct {
	DB *pgxpool.Pool
}

func NewProfileRepo(db *pgxpool.Pool) *ProfileRepo {
	return &ProfileRepo{DB: db}
}

func (r *ProfileRepo) CreateProfile(ctx context.Context, p models.UserProfile) error {
	logger.Info("ProfileRepo.CreateProfile", "creating profile", map[string]interface{}{"user_id": p.UserID})

	loc, _ := time.LoadLocation("Asia/Kolkata")
	p.CreatedAt = time.Now().In(loc)
	p.UpdatedAt = p.CreatedAt

	query := `
    INSERT INTO user_profiles (user_id, bio, avatar_url, location, dob, preferences, created_at, updated_at)
    VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
    `

	_, err := r.DB.Exec(ctx, query,
		p.UserID, p.Bio, p.AvatarURL, p.Location, p.DOB, p.Preferences, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		logger.Error("ProfileRepo.CreateProfile", "db insert failed", map[string]interface{}{"error": err.Error()})
		return fmt.Errorf("%w: %v", errors.ErrDatabaseFailure, err)
	}

	return nil
}

func (r *ProfileRepo) GetProfileByUserID(ctx context.Context, userID int) (*models.UserProfile, error) {
	logger.Info("ProfileRepo.GetProfileByUserID", "fetching profile", map[string]interface{}{"user_id": userID})

	query := `
    SELECT id, user_id, bio, avatar_url, location, dob, preferences, created_at, updated_at
    FROM user_profiles WHERE user_id=$1
    `

	var p models.UserProfile
	err := r.DB.QueryRow(ctx, query, userID).Scan(&p.ID, &p.UserID, &p.Bio, &p.AvatarURL, &p.Location, &p.DOB, &p.Preferences, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			logger.Warn("ProfileRepo.GetProfileByUserID", "profile not found", map[string]interface{}{"user_id": userID})
			return nil, errors.ErrResourceNotFound
		}
		logger.Error("ProfileRepo.GetProfileByUserID", "db error", map[string]interface{}{"error": err.Error()})
		return nil, fmt.Errorf("%w: %v", errors.ErrDatabaseFailure, err)
	}

	return &p, nil
}

func (r *ProfileRepo) UpdateProfile(ctx context.Context, p models.UserProfile) error {
	logger.Info("ProfileRepo.UpdateProfile", "updating profile", map[string]interface{}{"user_id": p.UserID})

	p.UpdatedAt = time.Now()

	query := `
    UPDATE user_profiles SET bio=$1, avatar_url=$2, location=$3, dob=$4, preferences=$5, updated_at=$6
    WHERE user_id=$7
    `

	val, err := r.DB.Exec(ctx, query, p.Bio, p.AvatarURL, p.Location, p.DOB, p.Preferences, p.UpdatedAt, p.UserID)
	if err != nil {
		logger.Error("ProfileRepo.UpdateProfile", "db update failed", map[string]interface{}{"error": err.Error()})
		return fmt.Errorf("%w: %v", errors.ErrDatabaseFailure, err)
	}

	if val.RowsAffected() == 0 {
		logger.Warn("ProfileRepo.UpdateProfile", "profile not found to update", map[string]interface{}{"user_id": p.UserID})
		return errors.ErrResourceNotFound
	}

	return nil
}
