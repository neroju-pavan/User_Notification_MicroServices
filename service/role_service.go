package service

import (
	"context"
	"errors"
	"test123/models"
	"test123/repositories"
)

type RoleService struct {
	RoleMod repositories.RoleRepoInter
}

func NewRoleService(rolemod repositories.RoleRepoInter) *RoleService {
	return &RoleService{
		RoleMod: rolemod,
	}
}

func (s *RoleService) CreateRole(ctx context.Context, name string) error {

	_, err := s.RoleMod.CreateRole(ctx, name)
	if err != nil {
		return errors.New("database failure")
	}

	return nil
}

func (s *RoleService) GetRoleByName(ctx context.Context, name string) (*models.Role, error) {

	u, err := s.RoleMod.GetRoleByName(ctx, name)
	if err != nil {
		return nil, errors.New("database failure")
	}

	return u, nil
}
