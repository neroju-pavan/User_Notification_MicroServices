package repositories

import (
	"context"

	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRoleRepo struct {
	DB *pgxpool.Pool
}

func NewUserRoleRepo(db *pgxpool.Pool) *UserRoleRepo {
	return &UserRoleRepo{
		DB: db,
	}
}

func (r *UserRoleRepo) AddUserRole(ctx context.Context, role string, user int) error {

	//get role id from db with name frm roles
	var roleID int
	err := r.DB.QueryRow(ctx,
		`SELECT id FROM roles WHERE name = $1`,
		role,
	).Scan(&roleID)

	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("role '%s' not found", role)
		}
		return err
	}

	//then add user_id,role_id into user_roles
	_, err = r.DB.Exec(ctx,
		`INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)`,
		user, roleID,
	)
	if err != nil {
		return err
	}

	return nil
}
