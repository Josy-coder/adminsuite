package database

import (
	"gorm.io/gorm"

	"github.com/josy-coder/adminsuite/internal/models"

)

func SeedDatabase(db *gorm.DB) error {
    // Create a default tenant
    tenant := models.Tenant{
        Name:   "Default Tenant",
        Domain: "default.adminsuite.com",
    }
    if err := db.Create(&tenant).Error; err != nil {
        return err
    }

    // Create default roles
    adminRole := models.Role{
        TenantID:    tenant.ID,
        Name:        "Admin",
        Description: "Administrator role",
    }
    userRole := models.Role{
        TenantID:    tenant.ID,
        Name:        "User",
        Description: "Regular user role",
    }
    if err := db.Create(&adminRole).Error; err != nil {
        return err
    }
    if err := db.Create(&userRole).Error; err != nil {
        return err
    }

    // Create a default admin user
    adminUser := models.User{
        TenantID:  tenant.ID,
        Email:     "admin@adminsuite.com",
        Username:  "admin",
        Password:  "adminpassword", 
        FirstName: "Admin",
        LastName:  "User",
        IsActive:  true,
    }
    if err := db.Create(&adminUser).Error; err != nil {
        return err
    }

    // Assign admin role to admin user
    if err := db.Model(&adminUser).Association("Roles").Append(&adminRole); err != nil {
        return err
    }

    return nil
}
