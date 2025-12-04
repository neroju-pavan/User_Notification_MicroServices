package repositories

import "context"

type UserRoleRepoInterface interface {
	AddUserRole(ctx context.Context, role string, user int) error
}
