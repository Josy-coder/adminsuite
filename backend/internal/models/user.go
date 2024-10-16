package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;" json:"id" swaggertype:"string" format:"uuid"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty" swaggertype:"string" format:"date-time"`
}

func (base *BaseModel) BeforeCreate(tx *gorm.DB) error {
	base.ID = uuid.New()
	return nil
}

// Tenant represents a tenant in the system
type Tenant struct {
	BaseModel
	Name             string `gorm:"size:255;not null" json:"name"`
	Domain           string `gorm:"size:255;uniqueIndex" json:"domain"`
	IsActive         bool   `gorm:"default:true" json:"is_active"`
	AuthPolicyConfig string `gorm:"type:jsonb" json:"auth_policy_config"`
	BrandingConfig   string `gorm:"type:jsonb" json:"branding_config"`
}

type MFAMethod string

const (
	MFAMethodTOTP  MFAMethod = "totp"
	MFAMethodSMS   MFAMethod = "sms"
	MFAMethodEmail MFAMethod = "email"
	MFAMethodHOTP  MFAMethod = "hotp"
)

type User struct {
	BaseModel
	TenantID           uuid.UUID `gorm:"type:uuid;index"`
	Email              string    `gorm:"size:255;uniqueIndex"`
	Username           string    `gorm:"size:50;uniqueIndex"`
	Password           string    `gorm:"size:255"`
	FirstName          string    `gorm:"size:50"`
	LastName           string    `gorm:"size:50"`
	PhoneNumber        string    `gorm:"size:20"`
	IsActive           bool      `gorm:"default:true"`
	EmailVerified      bool      `gorm:"default:false"`
	MFAEnabled         bool      `gorm:"default:false"`
	MFASecret          string    `gorm:"size:64"`
	MFAMethod          MFAMethod `gorm:"size:10"`
	MFABackupCodes     []string  `gorm:"type:json"`
	MFASMSCode         string    `gorm:"size:6"`
	MFASMSCodeExpiry   time.Time
	MFAEmailCode       string `gorm:"size:6"`
	MFAEmailCodeExpiry time.Time
	MFAHOTPCounter     uint64
	LastLoginAt        *time.Time
	PasswordChangedAt  *time.Time
	ProfilePicture     string `gorm:"size:255"`
	PreferencesConfig  string `gorm:"type:json"`
	Roles              []Role `gorm:"many2many:user_roles;"`
}

type Role struct {
	BaseModel
	TenantID    uuid.UUID    `gorm:"type:uuid;index"`
	Name        string       `gorm:"size:50"`
	Description string       `gorm:"size:255"`
	Permissions []Permission `gorm:"many2many:role_permissions;"`
}

type Permission struct {
	BaseModel
	TenantID    uuid.UUID `gorm:"type:uuid;index"`
	Name        string    `gorm:"size:50"`
	Description string    `gorm:"size:255"`
}

type TokenType string

const (
	TokenTypeRefresh       TokenType = "refresh"
	TokenTypeAccess        TokenType = "access"
	TokenTypePasswordReset TokenType = "password_reset"
)

type Token struct {
	BaseModel
	UserID    uuid.UUID `gorm:"type:uuid;index"`
	Token     string    `gorm:"size:255;uniqueIndex"`
	Type      TokenType `gorm:"size:20"`
	ExpiresAt time.Time
}

type PasswordReset struct {
	BaseModel
	UserID  uuid.UUID `gorm:"type:uuid;index"`
	TokenID uuid.UUID `gorm:"type:uuid;index"`
	Token   Token     `gorm:"foreignKey:TokenID"`
}

type LoginAttempt struct {
	BaseModel
	UserID    uuid.UUID `gorm:"type:uuid;index"`
	IP        string    `gorm:"size:45"`
	UserAgent string    `gorm:"size:255"`
	Success   bool
}

type AuditLog struct {
	BaseModel
	UserID   uuid.UUID `gorm:"type:uuid;index"`
	TenantID uuid.UUID `gorm:"type:uuid;index"`
	Action   string    `gorm:"size:50"`
	Resource string    `gorm:"size:50"`
	Details  string    `gorm:"type:text"`
}

type APIKey struct {
	BaseModel
	UserID      uuid.UUID `gorm:"type:uuid;index"`
	TenantID    uuid.UUID `gorm:"type:uuid;index"`
	Key         string    `gorm:"size:255;uniqueIndex"`
	Name        string    `gorm:"size:50"`
	Permissions string    `gorm:"type:jsonb"`
	ExpiresAt   *time.Time
}

type Device struct {
	BaseModel
	UserID     uuid.UUID `gorm:"type:uuid;index"`
	Name       string    `gorm:"size:50"`
	Type       string    `gorm:"size:20"`
	LastUsedAt time.Time
}
