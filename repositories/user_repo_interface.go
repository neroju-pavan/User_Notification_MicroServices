package repositories

import (
	"context"
	"test123/models"
	"time"
)

type UserRepoInterface interface {
	CreateUser(ctx context.Context, user models.User) error
	GetAllUsers(ctx context.Context) ([]models.User, error)
	UpdateUser(ctx context.Context, user models.User) error
	DeleteUser(ctx context.Context, id int) error
	GetUserByID(ctx context.Context, id int) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	UpdatePassword(ctx context.Context, email string, password string) error
	GetUserByEmailOrUsername(ctx context.Context, key string) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	GetUsersWithFiltersCursor(ctx context.Context, limit int, cursor *time.Time, usernameSearch string, fromDate, toDate *time.Time) ([]models.User, *time.Time, error)
}
