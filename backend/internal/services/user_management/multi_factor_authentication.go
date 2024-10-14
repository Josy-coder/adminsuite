package user_management

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"net/smtp"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/hotp"
	"github.com/pquerna/otp/totp"
	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"

	"github.com/josy-coder/adminsuite/internal/config"
	"github.com/josy-coder/adminsuite/internal/models"
	"github.com/josy-coder/adminsuite/internal/repositories/user_management"

)

type MFAService struct {
    userRepo user_management.UserRepository
    config   *config.Config
    twilio   *twilio.RestClient
}

func NewMFAService(userRepo user_management.UserRepository, config *config.Config) *MFAService {
    twilioClient := twilio.NewRestClientWithParams(twilio.ClientParams{
        Username: config.TwilioAccountSID,
        Password: config.TwilioAuthToken,
    })

    return &MFAService{
        userRepo: userRepo,
        config:   config,
        twilio:   twilioClient,
    }
}

func (s *MFAService) GenerateTOTPSecret(user *models.User) (string, error) {
    secret := make([]byte, 32)
    _, err := rand.Read(secret)
    if err != nil {
        return "", err
    }

    secretBase32 := base32.StdEncoding.EncodeToString(secret)
    user.MFASecret = secretBase32
    user.MFAEnabled = false
    user.MFAMethod = models.MFAMethodTOTP

    err = s.userRepo.Update(user)
    if err != nil {
        return "", err
    }

    return secretBase32, nil
}

func (s *MFAService) GenerateTOTPQRCode(user *models.User) (string, error) {
    if user.MFASecret == "" {
        return "", fmt.Errorf("MFA secret not set for user")
    }

    key, err := totp.Generate(totp.GenerateOpts{
        Issuer:      "AdminSuite",
        AccountName: user.Email,
        Secret:      []byte(user.MFASecret),
    })
    if err != nil {
        return "", err
    }

    return key.URL(), nil
}

func (s *MFAService) VerifyTOTP(user *models.User, token string) (bool, error) {
    if user.MFASecret == "" {
        return false, fmt.Errorf("MFA not set up for user")
    }

    valid := totp.Validate(token, user.MFASecret)
    if valid {
        user.MFAEnabled = true
        err := s.userRepo.Update(user)
        if err != nil {
            return true, fmt.Errorf("failed to update user MFA status: %v", err)
        }
    }

    return valid, nil
}

func (s *MFAService) GenerateBackupCodes(user *models.User) ([]string, error) {
    codes := make([]string, 10)
    for i := 0; i < 10; i++ {
        code, err := generateRandomCode(8)
        if err != nil {
            return nil, err
        }
        codes[i] = code
    }

    user.MFABackupCodes = codes
    err := s.userRepo.Update(user)
    if err != nil {
        return nil, err
    }

    return codes, nil
}

func (s *MFAService) VerifyBackupCode(user *models.User, code string) (bool, error) {
    for i, storedCode := range user.MFABackupCodes {
        if storedCode == code {
            // Remove the used backup code
            user.MFABackupCodes = append(user.MFABackupCodes[:i], user.MFABackupCodes[i+1:]...)
            err := s.userRepo.Update(user)
            if err != nil {
                return true, fmt.Errorf("failed to update user backup codes: %v", err)
            }
            return true, nil
        }
    }
    return false, nil
}

func (s *MFAService) GenerateSMSCode(user *models.User) (string, error) {
    code, err := generateRandomCode(6)
    if err != nil {
        return "", err
    }

    user.MFASMSCode = code
    user.MFASMSCodeExpiry = time.Now().Add(5 * time.Minute)
    user.MFAMethod = models.MFAMethodSMS
    err = s.userRepo.Update(user)
    if err != nil {
        return "", err
    }

    err = s.sendSMS(user.PhoneNumber, fmt.Sprintf("Your AdminSuite verification code is: %s", code))
    if err != nil {
        return "", err
    }

    return code, nil
}

func (s *MFAService) VerifySMSCode(user *models.User, code string) (bool, error) {
    if user.MFASMSCode == "" || time.Now().After(user.MFASMSCodeExpiry) {
        return false, fmt.Errorf("SMS code expired or not set")
    }

    if user.MFASMSCode == code {
        user.MFAEnabled = true
        user.MFASMSCode = ""
        user.MFASMSCodeExpiry = time.Time{}
        err := s.userRepo.Update(user)
        if err != nil {
            return true, fmt.Errorf("failed to update user MFA status: %v", err)
        }
        return true, nil
    }

    return false, nil
}

func (s *MFAService) GenerateEmailCode(user *models.User) (string, error) {
    code, err := generateRandomCode(6)
    if err != nil {
        return "", err
    }

    user.MFAEmailCode = code
    user.MFAEmailCodeExpiry = time.Now().Add(15 * time.Minute)
    user.MFAMethod = models.MFAMethodEmail
    err = s.userRepo.Update(user)
    if err != nil {
        return "", err
    }

    err = s.sendEmail(user.Email, "AdminSuite MFA Code", fmt.Sprintf("Your verification code is: %s", code))
    if err != nil {
        return "", err
    }

    return code, nil
}

func (s *MFAService) VerifyEmailCode(user *models.User, code string) (bool, error) {
    if user.MFAEmailCode == "" || time.Now().After(user.MFAEmailCodeExpiry) {
        return false, fmt.Errorf("email code expired or not set")
    }

    if user.MFAEmailCode == code {
        user.MFAEnabled = true
        user.MFAEmailCode = ""
        user.MFAEmailCodeExpiry = time.Time{}
        err := s.userRepo.Update(user)
        if err != nil {
            return true, fmt.Errorf("failed to update user MFA status: %v", err)
        }
        return true, nil
    }

    return false, nil
}

func (s *MFAService) GenerateHOTP(user *models.User) (string, error) {
    secret := make([]byte, 32)
    _, err := rand.Read(secret)
    if err != nil {
        return "", err
    }

    secretBase32 := base32.StdEncoding.EncodeToString(secret)
    user.MFASecret = secretBase32
    user.MFAEnabled = false
    user.MFAMethod = models.MFAMethodHOTP
    user.MFAHOTPCounter = 0

    err = s.userRepo.Update(user)
    if err != nil {
        return "", err
    }

    return secretBase32, nil
}

func (s *MFAService) VerifyHOTP(user *models.User, token string) (bool, error) {
	if user.MFASecret == "" {
		return false, fmt.Errorf("HOTP not set up for user")
	}

	secretBytes, err := base32.StdEncoding.DecodeString(user.MFASecret)
	if err != nil {
		return false, fmt.Errorf("failed to decode HOTP secret: %v", err)
	}

	valid, err := hotp.ValidateCustom(token, user.MFAHOTPCounter, string(secretBytes), hotp.ValidateOpts{
		Digits:    6,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		return false, fmt.Errorf("failed to validate HOTP: %v", err)
	}

	if valid {
		user.MFAEnabled = true
		user.MFAHOTPCounter++
		err := s.userRepo.Update(user)
		if err != nil {
			return true, fmt.Errorf("failed to update user MFA status: %v", err)
		}
	}

	return valid, nil
}

func (s *MFAService) sendSMS(phoneNumber, message string) error {
    params := &twilioApi.CreateMessageParams{}
    params.SetTo(phoneNumber)
    params.SetFrom(s.config.TwilioPhoneNumber)
    params.SetBody(message)

    _, err := s.twilio.Api.CreateMessage(params)
    return err
}

func (s *MFAService) sendEmail(to, subject, body string) error {
    auth := smtp.PlainAuth("", s.config.SMTPUsername, s.config.SMTPPassword, s.config.SMTPHost)

    msg := []byte(fmt.Sprintf("To: %s\r\n"+
        "From: %s\r\n"+
        "Subject: %s\r\n"+
        "\r\n"+
        "%s\r\n", to, s.config.SMTPFrom, subject, body))

    err := smtp.SendMail(fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort),
        auth,
        s.config.SMTPFrom,
        []string{to},
        msg)

    return err
}

func generateRandomCode(length int) (string, error) {
    const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    code := make([]byte, length)
    _, err := rand.Read(code)
    if err != nil {
        return "", err
    }
    for i := range code {
        code[i] = charset[int(code[i])%len(charset)]
    }
    return string(code), nil
}
