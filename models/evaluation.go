package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Evaluation struct {
	ID            uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	InstitutionID uuid.UUID      `json:"institution_id" gorm:"type:uuid;not null"`
	Institution   Institution    `json:"institution" gorm:"foreignKey:InstitutionID"`
	UserID        uuid.UUID      `json:"user_id" gorm:"type:uuid"`
	User          *User          `json:"user,omitempty" gorm:"foreignKey:UserID"`
	IsAnonymous   bool           `json:"is_anonymous" gorm:"default:false"`
	Rating        int            `json:"rating" gorm:"not null;check:rating >= 1 AND rating <= 5"`
	Transparency  int            `json:"transparency" gorm:"check:transparency >= 1 AND transparency <= 5"`
	Accountability int           `json:"accountability" gorm:"check:accountability >= 1 AND accountability <= 5"`
	Responsiveness int           `json:"responsiveness" gorm:"check:responsiveness >= 1 AND responsiveness <= 5"`
	Integrity     int            `json:"integrity" gorm:"check:integrity >= 1 AND integrity <= 5"`
	Comment       string         `json:"comment" gorm:"type:text"`
	Province      string         `json:"province"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

func (e *Evaluation) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}
