package user_management

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/o1egl/paseto"
	"golang.org/x/crypto/argon2"

	"github.com/josy-coder/adminsuite/internal/models"
	"github.com/josy-coder/adminsuite/internal/repositories/user_management"

)

var ErrMFARequired = errors.New("MFA required")

type AuthenticationService struct {
	userRepo  user_management.UserRepository
	tokenRepo user_management.TokenRepository
	paseto    paseto.V2
	pasetoKey []byte
	mfaService *MFAService
}

func NewAuthenticationService(
	userRepo user_management.UserRepository,
	tokenRepo user_management.TokenRepository,
	pasetoKey []byte,
	mfaService *MFAService,
) *AuthenticationService {
	return &AuthenticationService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		paseto:    paseto.V2{},
		pasetoKey: pasetoKey,
		mfaService: mfaService,
	}
}

type argon2Params struct {
	memory uint32
	iterations uint32
	parallelism uint8
	saltLength uint32
	keyLength uint32
}

var params = &argon2Params{
	memory: 64 * 1024,
	iterations: 3,
	parallelism: 4,
	saltLength: 16,
	keyLength: 32,
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

	match, err := s.verifyPassword(user.Password, password)
	if err != nil {
		return nil, "", "", errors.New("error verifying password")
	}
	if !match {
		return nil, "","", errors.New("invalid credentials")
	}

	if user.MFAEnabled {
		return user, "", "", errors.New("multi factor authentication required")
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
	salt, err := generateRandomBytes(params.saltLength)
	if err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, params.iterations, params.memory, params.parallelism, params.keyLength)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, params.memory, params.iterations, params.parallelism, b64Salt, b64Hash)

	return encodedHash, nil
}

func (s *AuthenticationService) verifyPassword(encodedhash, password string) (bool, error) {
	p, salt, hash, err := decodeHash(encodedhash)
	if err!= nil {
        return false, err
    }

	otherHash := argon2.IDKey([]byte(password), salt, p.iterations, p.memory, p.parallelism, p.keyLength)

	return base64.RawStdEncoding.EncodeToString(hash) == base64.RawStdEncoding.EncodeToString(otherHash), nil
}

func generateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err!= nil {
        return nil, err
    }
	return b, nil
}

func decodeHash(encodedHash string) (p *argon2Params, salt, hash []byte, err error) {
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return nil, nil, nil, errors.New("invalid hash format")
	}

	var version int
	_, err = fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, err
	}
	if version != argon2.Version {
		return nil, nil, nil, fmt.Errorf("incompatible version")
	}

	p = &argon2Params{}
	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &p.memory, &p.iterations, &p.parallelism)
	if err!= nil {
        return nil, nil, nil, err
    }

	salt, err = base64.RawStdEncoding.DecodeString(vals[4])
	if err!= nil {
        return nil, nil, nil, err
    }

	p.keyLength = uint32(len(hash))

	return p, salt, hash, nil
}

func (s *AuthenticationService) ValidateToken(token string) (*paseto.JSONToken, error) {
    var claims paseto.JSONToken
    err := s.paseto.Decrypt(token, s.pasetoKey, &claims, nil)
    if err != nil {
        return nil, err
    }
    return &claims, nil
}

func (s *AuthenticationService) GetUserByID(userID string) (*models.User, error) {
    id, err := uuid.Parse(userID)
    if err != nil {
        return nil, err
    }
    return s.userRepo.FindByID(id)
}

func (s *AuthenticationService) GenerateTempToken(userID uuid.UUID) (string, error) {
	now := time.Now()
	exp := now.Add(5 * time.Minute)
	nbt := now

	token := paseto.JSONToken{
		Audience:   "adminsuite-mfa",
		Issuer:     "adminsuite-auth",
		Jti:        uuid.New().String(),
		Subject:    userID.String(),
		IssuedAt:   now,
		Expiration: exp,
		NotBefore:  nbt,
	}

	return s.paseto.Encrypt(s.pasetoKey, token, nil)
}

func (s *AuthenticationService) GetUserByTempToken(tempToken string) (*models.User, error) {
	var token paseto.JSONToken
	err := s.paseto.Decrypt(tempToken, s.pasetoKey, &token, nil)
	if err != nil {
		return nil, err
	}

	if time.Now().After(token.Expiration) {
		return nil, errors.New("temporary token expired")
	}

	userID, err := uuid.Parse(token.Subject)
	if err != nil {
		return nil, err
	}

	return s.userRepo.FindByID(userID)
}

func (s *AuthenticationService) GenerateTokens(user *models.User) (string, string, error) {
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := s.generateRefreshToken(user)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
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