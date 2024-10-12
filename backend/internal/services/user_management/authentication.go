package user_management

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/o1egl/paseto"
	"golang.org/x/crypto/argon2"

	"github.com/josy-coder/adminsuite/internal/models"
	"github.com/josy-coder/adminsuite/internal/repositories/user_management"

)

type AuthenticationService struct {
	userRepo  user_management.UserRepository
	tokenRepo user_management.TokenRepository
	paseto    paseto.V2
	pasetoKey []byte
}

func NewAuthenticationService(
	userRepo user_management.UserRepository,
	tokenRepo user_management.TokenRepository,
	pasetoKey []byte,
) *AuthenticationService {
	return &AuthenticationService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		paseto:    paseto.V2{},
		pasetoKey: pasetoKey,
	}
}

func (s *AuthenticationService) RegisterUser(user *models.User) error {
	hashedPassword, err := s.hashPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = hashedPassword

	return s.userRepo.Create(user)
}

func (s *AuthenticationService) AuthenticateUser(email, password string) (*models.User, string, string, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, "", "", errors.New("invalid credentials")
	}

	if !s.verifyPassword(user.Password, password) {
		return nil, "", "", errors.New("invalid credentials")
	}

	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, "", "", err
	}

	refreshToken, err := s.generateRefreshToken(user)
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

func (s *AuthenticationService) RefreshToken(refreshToken string) (string, string, error) {
	tokenData, err := s.tokenRepo.FindByToken(refreshToken)
	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}

	if tokenData.ExpiresAt.Before(time.Now()) {
		return "", "", errors.New("refresh token expired")
	}

	user, err := s.userRepo.FindByID(tokenData.UserID)
	if err != nil {
		return "", "", errors.New("user not found")
	}

	newAccessToken, err := s.generateAccessToken(user)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := s.generateRefreshToken(user)
	if err != nil {
		return "", "", err
	}

	if err := s.tokenRepo.Delete(tokenData.ID); err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

func (s *AuthenticationService) hashPassword(password string) (string, error) {
	hash := argon2.IDKey([]byte(password), []byte("salt"), 1, 64*1024, 4, 32)
	return string(hash), nil
}

func (s *AuthenticationService) verifyPassword(hashedPassword, password string) bool {
	hash := argon2.IDKey([]byte(password), []byte("salt"), 1, 64*1024, 4, 32)
	return string(hash) == hashedPassword
}

func (s *AuthenticationService) generateAccessToken(user *models.User) (string, error) {
	now := time.Now()
	exp := now.Add(15 * time.Minute)
	nbt := now

	token := paseto.JSONToken{
		Audience:   "adminsuite",
		Issuer:     "adminsuite-auth",
		Jti:        uuid.New().String(),
		Subject:    user.ID.String(),
		IssuedAt:   now,
		Expiration: exp,
		NotBefore:  nbt,
	}

	return s.paseto.Encrypt(s.pasetoKey, token, nil)
}

func (s *AuthenticationService) generateRefreshToken(user *models.User) (string, error) {
	token := &models.Token{
		UserID:    user.ID,
		Token:     uuid.New().String(),
		Type:      models.TokenTypeRefresh,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	if err := s.tokenRepo.Create(token); err != nil {
		return "", err
	}

	return token.Token, nil
}