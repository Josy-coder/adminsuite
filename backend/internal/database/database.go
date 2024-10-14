package database

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/josy-coder/adminsuite/internal/config"
	"github.com/josy-coder/adminsuite/internal/models"
)

func SetupDatabase(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)

	log.Printf("Connecting to database with DSN: %s", dsn)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	log.Println("Database connected successfully")

	err = runMigrations(db)
	if err != nil {
		return nil, fmt.Errorf("failed to run migrations: %v", err)
	}

	return db, nil
}

func runMigrations(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.Tenant{},
		&models.User{},
		&models.Role{},
		&models.Permission{},
		&models.Token{},
		&models.PasswordReset{},
		&models.LoginAttempt{},
		&models.AuditLog{},
		&models.APIKey{},
		&models.Device{},
	)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	log.Println("Migrations completed successfully")
	return nil
}
