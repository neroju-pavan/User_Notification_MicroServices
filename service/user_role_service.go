package service

import (
	"context"
	"test123/repositories"
)

type UserRoleService struct {
	Repo repositories.UserRoleRepoInterface
}

func NewUserRoleService(repo repositories.UserRoleRepoInterface) *UserRoleService {
	return &UserRoleService{
		Repo: repo,
	}
}

func (s *UserRoleService) AddUserRole(context context.Context, role string, d int) error {
	return s.Repo.AddUserRole(context, role, d)

}
