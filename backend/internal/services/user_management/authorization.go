package user_management

import (
	"errors"

	"github.com/google/uuid"

	"github.com/josy-coder/adminsuite/internal/models"
	"github.com/josy-coder/adminsuite/internal/repositories/user_management"

)

type AuthorizationService struct {
	userRepo       user_management.UserRepository
	roleRepo       user_management.RoleRepository
	permissionRepo user_management.PermissionRepository
}

func NewAuthorizationService(
	userRepo user_management.UserRepository,
	roleRepo user_management.RoleRepository,
	permissionRepo user_management.PermissionRepository,
) *AuthorizationService {
	return &AuthorizationService{
		userRepo:       userRepo,
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
	}
}

func (s *AuthorizationService) AssignRoleToUser(userID, roleID string) error {
	user, err := s.userRepo.FindByID(uuid.MustParse(userID))
	if err != nil {
		return errors.New("user not found")
	}

	role, err := s.roleRepo.FindByID(uuid.MustParse(roleID))
	if err != nil {
		return errors.New("role not found")
	}

	user.Roles = append(user.Roles, *role)
	return s.userRepo.Update(user)
}

func (s *AuthorizationService) CheckUserPermission(userID, permissionName string) (bool, error) {
	user, err := s.userRepo.FindByID(uuid.MustParse(userID))
	if err != nil {
		return false, errors.New("user not found")
	}

	for _, role := range user.Roles {
		for _, permission := range role.Permissions {
			if permission.Name == permissionName {
				return true, nil
			}
		}
	}

	return false, nil
}

func (s *AuthorizationService) CreateRole(role *models.Role) error {
	return s.roleRepo.Create(role)
}

func (s *AuthorizationService) CreatePermission(permission *models.Permission) error {
	return s.permissionRepo.Create(permission)
}

func (s *AuthorizationService) AssignPermissionToRole(roleID, permissionID string) error {
	role, err := s.roleRepo.FindByID(uuid.MustParse(roleID))
	if err != nil {
		return errors.New("role not found")
	}

	permission, err := s.permissionRepo.FindByID(uuid.MustParse(permissionID))
	if err != nil {
		return errors.New("permission not found")
	}

	role.Permissions = append(role.Permissions, *permission)
	return s.roleRepo.Update(role)
}
