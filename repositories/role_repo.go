package repositories

import (
	"context"
	"test123/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RoleRepo struct {
	DB *pgxpool.Pool
}

func NewRoleRepo(db *pgxpool.Pool) *RoleRepo {
	return &RoleRepo{DB: db}
}

func (r *RoleRepo) CreateRole(ctx context.Context, name string) (*models.Role, error) {
	var role models.Role
	query := `INSERT INTO roles (name) values ($1) RETURNING id,name`
	err := r.DB.QueryRow(ctx, query,name).Scan(&role.Id, &role.Name)
	return &role, err

}

func (r *RoleRepo) GetRoleByName(ctx context.Context, name string) (*models.Role, error) {

	var role models.Role
	query := `SELECT id,name FROM roles WHERE name=$1`
	err := r.DB.QueryRow(ctx, query).Scan(&role.Id, &role.Name)

	return &role, err
}
