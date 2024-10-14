package main

import (
	"log"
	// Import the docs package

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/josy-coder/adminsuite/docs/api"
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

	// Connect to database and run migrations
	db, err := database.SetupDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to setup database: %v", err)
	}

	// Seed the database
	if err := database.SeedDatabase(db); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	// Initialize repositories
	userRepo := user_management.NewUserRepository(db)
	tokenRepo := user_management.NewTokenRepository(db)

	// Initialize services
	mfaService := services.NewMFAService(userRepo, cfg)
	authService := services.NewAuthenticationService(userRepo, tokenRepo, cfg.PasetoKey, mfaService)

	// Initialize Gin router
	r := gin.Default()

	// Setup routes
	routes.SetupRoutes(r, authService, mfaService, userRepo)

	// Swagger route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start server
	log.Printf("Starting server on port %s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
