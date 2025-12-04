package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type AuthorizeService struct {
	DB    *pgxpool.Pool
	Cache *redis.Client
}

func NewAuthorizeService(db *pgxpool.Pool, rdb *redis.Client) *AuthorizeService {
	return &AuthorizeService{
		DB:    db,
		Cache: rdb,
	}
}

func (a *AuthorizeService) GetPermissions(ctx context.Context, userID int) ([]string, error) {

	query := `
    SELECT DISTINCT p.name
    FROM user_roles ur
    JOIN role_permissions rp ON ur.role_id = rp.role_id
    JOIN permissions p ON rp.permission_id = p.id
    WHERE ur.user_id = $1
`

	rows, err := a.DB.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []string
	for rows.Next() {
		var p string
		rows.Scan(&p)
		perms = append(perms, p)
	}

	return perms, nil
}
