package user_management

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/josy-coder/adminsuite/internal/models"
	"github.com/josy-coder/adminsuite/internal/services/user_management"

)

type AuthenticationHandler struct {
	authService *user_management.AuthenticationService
	mfaService  *user_management.MFAService
}

func NewAuthenticationHandler(authService *user_management.AuthenticationService, mfaService *user_management.MFAService) *AuthenticationHandler {
	return &AuthenticationHandler{
		authService: authService,
		mfaService:  mfaService,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user with the provided details
// @Tags authentication
// @Accept json
// @Produce json
// @Param user body RegisterRequest true "User Registration Details"
// @Success 201 {object} models.User
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/register [post]
func (h *AuthenticationHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := &models.User{
		Email:     req.Email,
		Username:  req.Username,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	if err := h.authService.RegisterUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// Login godoc
// @Summary Authenticate a user
// @Description Authenticate a user with email and password
// @Tags authentication
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "User Login Credentials"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/login [post]
func (h *AuthenticationHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, accessToken, refreshToken, err := h.authService.AuthenticateUser(req.Email, req.Password)
	if err != nil {
		if err == user_management.ErrMFARequired {
			tempToken, err := h.authService.GenerateTempToken(user.ID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate temporary token"})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"mfa_required": true,
				"mfa_method":   user.MFAMethod,
				"temp_token":   tempToken,
			})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		User: UserResponse{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Refresh the access token using a valid refresh token
// @Tags authentication
// @Accept json
// @Produce json
// @Param refresh_token header string true "Refresh Token"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/refresh [post]
func (h *AuthenticationHandler) RefreshToken(c *gin.Context) {
	refreshToken := c.GetHeader("Refresh-Token")
	if refreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Refresh token is required"})
		return
	}

	newAccessToken, newRefreshToken, err := h.authService.RefreshToken(refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	c.JSON(http.StatusOK, TokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	})
}

// VerifyMFA godoc
// @Summary Verify MFA token
// @Description Verify the MFA token provided by the user
// @Tags authentication
// @Accept json
// @Produce json
// @Param mfa_verification body MFAVerificationRequest true "MFA Verification Details"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/verify-mfa [post]
func (h *AuthenticationHandler) VerifyMFA(c *gin.Context) {
	var req MFAVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.GetUserByTempToken(req.TempToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid temporary token"})
		return
	}

	var valid bool
	switch user.MFAMethod {
	case models.MFAMethodTOTP:
		valid, err = h.mfaService.VerifyTOTP(user, req.MFAToken)
	case models.MFAMethodSMS:
		valid, err = h.mfaService.VerifySMSCode(user, req.MFAToken)
	case models.MFAMethodEmail:
		valid, err = h.mfaService.VerifyEmailCode(user, req.MFAToken)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid MFA method"})
		return
	}

	if err != nil || !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid MFA token"})
		return
	}

	accessToken, refreshToken, err := h.authService.GenerateTokens(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		User: UserResponse{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type MFAVerificationRequest struct {
	TempToken string `json:"temp_token" binding:"required"`
	MFAToken  string `json:"mfa_token" binding:"required"`
}

type LoginResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
}

type UserResponse struct {
	ID       uuid.UUID `json:"id"`
	Email    string    `json:"email"`
	Username string    `json:"username"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
