package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CategoryType string

const (
	CategoryProposal CategoryType = "proposal"
	CategoryReport   CategoryType = "report"
)

type Category struct {
	ID          uuid.UUID    `json:"id" gorm:"type:uuid;primaryKey"`
	Name        string       `json:"name" gorm:"not null"`
	Description string       `json:"description"`
	Icon        string       `json:"icon"`
	Color       string       `json:"color"`
	Type        CategoryType `json:"type" gorm:"type:varchar(20);not null"`
	IsActive    bool         `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

func (c *Category) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
