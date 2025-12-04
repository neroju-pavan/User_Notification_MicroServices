package repositories

import (
	"context"
	"test123/models"
)

type RoleRepoInter interface {
	CreateRole(ctx context.Context,name string)(*models.Role,error)
	GetRoleByName(ctx context.Context,name string)(*models.Role,error)
}