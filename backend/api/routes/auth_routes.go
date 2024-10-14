package routes

import (
	"github.com/gin-gonic/gin"

	handlers "github.com/josy-coder/adminsuite/api/handlers/user_management"
	"github.com/josy-coder/adminsuite/api/middleware"
	"github.com/josy-coder/adminsuite/internal/repositories/user_management"
	services "github.com/josy-coder/adminsuite/internal/services/user_management"

)

func SetupRoutes(r *gin.Engine, authService *services.AuthenticationService, mfaService *services.MFAService, userRepo user_management.UserRepository) {
	authHandler := handlers.NewAuthenticationHandler(authService, mfaService)
	mfaHandler := handlers.NewMFAHandler(authService, mfaService, userRepo)

	auth := r.Group("/api/v1/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)
		auth.POST("/verify-mfa", authHandler.VerifyMFA)
	}

	mfa := r.Group("/api/v1/mfa")
	mfa.Use(middleware.AuthMiddleware(authService))
	{
		mfa.POST("/setup/totp", mfaHandler.SetupTOTP)
		mfa.POST("/verify/totp", mfaHandler.VerifyTOTP)
		mfa.POST("/setup/sms", mfaHandler.SetupSMS)
		mfa.POST("/verify/sms", mfaHandler.VerifySMS)
		mfa.POST("/backup-codes", mfaHandler.GenerateBackupCodes)
		mfa.POST("/verify/backup", mfaHandler.VerifyBackupCode)
		mfa.POST("/disable", mfaHandler.DisableMFA)
	}
}