package main

import (
	"log"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/josy-coder/adminsuite/api/routes"
	"github.com/josy-coder/adminsuite/internal/config"
	"github.com/josy-coder/adminsuite/internal/database"
	"github.com/josy-coder/adminsuite/internal/repositories/user_management"
	services "github.com/josy-coder/adminsuite/internal/services/user_management"
)

// @title           AdminSuite API
// @version         1.0
// @description     This is the API for the AdminSuite application.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.adminsuite.com/support
// @contact.email  support@adminsuite.com

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database
	db, err := database.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize repositories
	userRepo := user_management.NewUserRepository(db)
	tokenRepo := user_management.NewTokenRepository(db)

	// Initialize services
	mfaService := services.NewMFAService(userRepo, cfg)
	authService := services.NewAuthenticationService(userRepo, tokenRepo, []byte(cfg.PasetoKey), mfaService)

	// Initialize Gin router
	r := gin.Default()

	// Setup routes
	routes.SetupRoutes(r, authService, mfaService, userRepo)

	// Swagger route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start server
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
