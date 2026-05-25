package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role string

const (
	RoleCitizen        Role = "citizen"
	RoleModerator      Role = "moderator"
	RoleRepresentative Role = "representative"
	RoleAdmin          Role = "admin"
	RoleSuperAdmin     Role = "super_admin"
)

type User struct {
	ID            uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	FullName      string         `json:"full_name" gorm:"not null"`
	Email         string         `json:"email" gorm:"uniqueIndex;not null"`
	Phone         string         `json:"phone" gorm:"uniqueIndex"`
	Password      string         `json:"-" gorm:"not null"`
	Role          Role           `json:"role" gorm:"type:varchar(20);default:'citizen'"`
	Province      string         `json:"province"`
	City          string         `json:"city"`
	Neighborhood  string         `json:"neighborhood"`
	Bio           string         `json:"bio"`
	Avatar        string         `json:"avatar"`
	IsVerified    bool           `json:"is_verified" gorm:"default:false"`
	IsActive      bool           `json:"is_active" gorm:"default:true"`
	PrefersSMS    bool           `json:"prefers_sms" gorm:"default:false"`
	LastLoginAt   *time.Time     `json:"last_login_at"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
