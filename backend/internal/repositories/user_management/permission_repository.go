package user_management

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/josy-coder/adminsuite/internal/models"

)

type PermissionRepository interface {
	Create(permission *models.Permission) error
	FindByID(id uuid.UUID) (*models.Permission, error)
	FindAll() ([]*models.Permission, error)
	Update(permission *models.Permission) error
	Delete(id uuid.UUID) error
}

type permissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) PermissionRepository {
	return &permissionRepository{db: db}
}

func (r *permissionRepository) Create(permission *models.Permission) error {
	return r.db.Create(permission).Error
}

func (r *permissionRepository) FindByID(id uuid.UUID) (*models.Permission, error) {
	var permission models.Permission
	err := r.db.First(&permission, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *permissionRepository) FindAll() ([]*models.Permission, error) {
	var permissions []*models.Permission
	err := r.db.Find(&permissions).Error
	return permissions, err
}

func (r *permissionRepository) Update(permission *models.Permission) error {
	return r.db.Save(permission).Error
}

func (r *permissionRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Permission{}, "id = ?", id).Error
}
