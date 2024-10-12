package user_management

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/josy-coder/adminsuite/internal/models"
)

type RoleRepository interface {
	Create(role *models.Role) error
	FindByID(id uuid.UUID) (*models.Role, error)
	FindAll() ([]*models.Role, error)
	Update(role *models.Role) error
	Delete(id uuid.UUID) error
}

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) Create(role *models.Role) error {
	return r.db.Create(role).Error
}

func (r *roleRepository) FindByID(id uuid.UUID) (*models.Role, error) {
	var role models.Role
	err := r.db.Preload("Permissions").First(&role, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) FindAll() ([]*models.Role, error) {
	var roles []*models.Role
	err := r.db.Preload("Permissions").Find(&roles).Error
	return roles, err
}

func (r *roleRepository) Update(role *models.Role) error {
	return r.db.Save(role).Error
}

func (r *roleRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Role{}, "id = ?", id).Error
}
