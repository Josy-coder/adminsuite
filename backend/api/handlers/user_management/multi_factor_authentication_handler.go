package user_management

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/josy-coder/adminsuite/internal/models"
	repositories "github.com/josy-coder/adminsuite/internal/repositories/user_management"
	services "github.com/josy-coder/adminsuite/internal/services/user_management"
)

type MFAHandler struct {
	authService *services.AuthenticationService
	mfaService  *services.MFAService
	userRepo    repositories.UserRepository
}

func NewMFAHandler(authService *services.AuthenticationService, mfaService *services.MFAService, userRepo repositories.UserRepository) *MFAHandler {
	return &MFAHandler{
		authService: authService,
		mfaService:  mfaService,
		userRepo:    userRepo,
	}
}

// SetupTOTP godoc
// @Summary Set up TOTP-based MFA
// @Description Generate a TOTP secret and QR code for the user
// @Tags MFA
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} TOTPSetupResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /mfa/setup/totp [post]
func (h *MFAHandler) SetupTOTP(c *gin.Context) {
	user := c.MustGet("user").(*models.User)

	secret, err := h.mfaService.GenerateTOTPSecret(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate MFA secret"})
		return
	}

	qrCode, err := h.mfaService.GenerateTOTPQRCode(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate QR code"})
		return
	}

	c.JSON(http.StatusOK, TOTPSetupResponse{
		Secret: secret,
		QRCode: qrCode,
	})
}

// VerifyTOTP godoc
// @Summary Verify TOTP-based MFA
// @Description Verify the TOTP token provided by the user
// @Tags MFA
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param token body TOTPVerificationRequest true "TOTP token"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /mfa/verify/totp [post]
func (h *MFAHandler) VerifyTOTP(c *gin.Context) {
	user := c.MustGet("user").(*models.User)

	var req TOTPVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	valid, err := h.mfaService.VerifyTOTP(user, req.Token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify MFA token"})
		return
	}
	if !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid MFA token"})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Message: "MFA verified successfully"})
}

// SetupSMS godoc
// @Summary Set up SMS-based MFA
// @Description Generate and send an SMS code for MFA
// @Tags MFA
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /mfa/setup/sms [post]
func (h *MFAHandler) SetupSMS(c *gin.Context) {
	user := c.MustGet("user").(*models.User)

	_, err := h.mfaService.GenerateSMSCode(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate SMS code"})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Message: "SMS code sent successfully"})
}

// VerifySMS godoc
// @Summary Verify SMS-based MFA
// @Description Verify the SMS code provided by the user
// @Tags MFA
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param code body SMSVerificationRequest true "SMS code"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /mfa/verify/sms [post]
func (h *MFAHandler) VerifySMS(c *gin.Context) {
	user := c.MustGet("user").(*models.User)

	var req SMSVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	valid, err := h.mfaService.VerifySMSCode(user, req.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify SMS code"})
		return
	}
	if !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid SMS code"})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Message: "SMS code verified successfully"})
}

// GenerateBackupCodes godoc
// @Summary Generate backup codes
// @Description Generate new backup codes for the user
// @Tags MFA
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} BackupCodesResponse
// @Failure 500 {object} ErrorResponse
// @Router /mfa/backup-codes [post]
func (h *MFAHandler) GenerateBackupCodes(c *gin.Context) {
	user := c.MustGet("user").(*models.User)

	codes, err := h.mfaService.GenerateBackupCodes(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate backup codes"})
		return
	}

	c.JSON(http.StatusOK, BackupCodesResponse{BackupCodes: codes})
}

// VerifyBackupCode godoc
// @Summary Verify backup code
// @Description Verify a backup code provided by the user
// @Tags MFA
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param code body BackupCodeVerificationRequest true "Backup code"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /mfa/verify/backup [post]
func (h *MFAHandler) VerifyBackupCode(c *gin.Context) {
	user := c.MustGet("user").(*models.User)

	var req BackupCodeVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	valid, err := h.mfaService.VerifyBackupCode(user, req.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify backup code"})
		return
	}
	if !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid backup code"})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Message: "Backup code verified successfully"})
}

// DisableMFA godoc
// @Summary Disable MFA
// @Description Disable MFA for the user
// @Tags MFA
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SuccessResponse
// @Failure 500 {object} ErrorResponse
// @Router /mfa/disable [post]
func (h *MFAHandler) DisableMFA(c *gin.Context) {
	user := c.MustGet("user").(*models.User)

	user.MFAEnabled = false
	user.MFASecret = ""
	user.MFAMethod = ""
	user.MFABackupCodes = nil

	if err := h.userRepo.Update(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to disable MFA"})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Message: "MFA disabled successfully"})
}

type TOTPSetupResponse struct {
	Secret string `json:"secret"`
	QRCode string `json:"qr_code"`
}

type TOTPVerificationRequest struct {
	Token string `json:"token" binding:"required"`
}

type SMSVerificationRequest struct {
	Code string `json:"code" binding:"required"`
}

type BackupCodesResponse struct {
	BackupCodes []string `json:"backup_codes"`
}

type BackupCodeVerificationRequest struct {
	Code string `json:"code" binding:"required"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}
