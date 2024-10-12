
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

)

type BaseModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (base *BaseModel) BeforeCreate(tx *gorm.DB) error {
	base.ID = uuid.New()
	return nil
}

type Tenant struct {
	BaseModel
	Name             string `gorm:"size:255;not null"`
	Domain           string `gorm:"size:255;uniqueIndex"`
	IsActive         bool   `gorm:"default:true"`
	AuthPolicyConfig string `gorm:"type:jsonb"`
	BrandingConfig   string `gorm:"type:jsonb"`
}

type User struct {
	BaseModel
	TenantID         uuid.UUID `gorm:"type:uuid;index"`
	Email            string    `gorm:"size:255;uniqueIndex"`
	Username         string    `gorm:"size:50;uniqueIndex"`
	Password         string    `gorm:"size:255"`
	FirstName        string    `gorm:"size:50"`
	LastName         string    `gorm:"size:50"`
	IsActive         bool      `gorm:"default:true"`
	EmailVerified    bool      `gorm:"default:false"`
	MFAEnabled       bool      `gorm:"default:false"`
	MFASecret        string    `gorm:"size:32"`
	LastLoginAt      *time.Time
	PasswordChangedAt *time.Time
	ProfilePicture   string    `gorm:"size:255"`
	PreferencesConfig string   `gorm:"type:jsonb"`
	Roles            []Role    `gorm:"many2many:user_roles;"`
}

type Role struct {
	BaseModel
	TenantID    uuid.UUID `gorm:"type:uuid;index"`
	Name        string    `gorm:"size:50"`
	Description string    `gorm:"size:255"`
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
	TokenTypeRefresh TokenType = "refresh"
	TokenTypeAccess  TokenType = "access"
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
	UserID    uuid.UUID `gorm:"type:uuid;index"`
	TokenID   uuid.UUID `gorm:"type:uuid;index"`
	Token     Token     `gorm:"foreignKey:TokenID"`
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
	UserID    uuid.UUID `gorm:"type:uuid;index"`
	TenantID  uuid.UUID `gorm:"type:uuid;index"`
	Action    string    `gorm:"size:50"`
	Resource  string    `gorm:"size:50"`
	Details   string    `gorm:"type:text"`
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
	UserID    uuid.UUID `gorm:"type:uuid;index"`
	Name      string    `gorm:"size:50"`
	Type      string    `gorm:"size:20"`
	LastUsedAt time.Time
}