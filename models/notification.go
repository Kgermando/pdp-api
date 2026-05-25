package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NotificationType string

const (
	NotifProposal    NotificationType = "proposal"
	NotifReport      NotificationType = "report"
	NotifEvaluation  NotificationType = "evaluation"
	NotifSystem      NotificationType = "system"
	NotifAlert       NotificationType = "alert"
)

type Notification struct {
	ID         uuid.UUID        `json:"id" gorm:"type:uuid;primaryKey"`
	UserID     uuid.UUID        `json:"user_id" gorm:"type:uuid;not null"`
	Title      string           `json:"title" gorm:"not null"`
	Message    string           `json:"message" gorm:"type:text;not null"`
	Type       NotificationType `json:"type" gorm:"type:varchar(20)"`
	RefID      string           `json:"ref_id"`
	IsRead     bool             `json:"is_read" gorm:"default:false"`
	CreatedAt  time.Time        `json:"created_at"`
	DeletedAt  gorm.DeletedAt   `json:"-" gorm:"index"`
}

func (n *Notification) BeforeCreate(tx *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return nil
}

type SMSLog struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	Phone     string    `json:"phone" gorm:"not null"`
	Message   string    `json:"message" gorm:"type:text;not null"`
	Status    string    `json:"status" gorm:"default:'pending'"`
	Provider  string    `json:"provider"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *SMSLog) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}
