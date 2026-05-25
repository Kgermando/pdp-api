package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type InstitutionType string

const (
	InstitutionMinistry    InstitutionType = "ministry"
	InstitutionParliament  InstitutionType = "parliament"
	InstitutionProvincial  InstitutionType = "provincial"
	InstitutionMunicipality InstitutionType = "municipality"
	InstitutionJudiciary   InstitutionType = "judiciary"
	InstitutionOther       InstitutionType = "other"
)

type Institution struct {
	ID           uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey"`
	Name         string          `json:"name" gorm:"not null"`
	Type         InstitutionType `json:"type" gorm:"type:varchar(20);not null"`
	Description  string          `json:"description" gorm:"type:text"`
	Province     string          `json:"province"`
	City         string          `json:"city"`
	Website      string          `json:"website"`
	Phone        string          `json:"phone"`
	Email        string          `json:"email"`
	Address      string          `json:"address"`
	Logo         string          `json:"logo"`
	IsActive     bool            `json:"is_active" gorm:"default:true"`
	AvgRating    float64         `json:"avg_rating" gorm:"default:0"`
	TotalRatings int             `json:"total_ratings" gorm:"default:0"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	DeletedAt    gorm.DeletedAt  `json:"-" gorm:"index"`
	Evaluations  []Evaluation    `json:"evaluations,omitempty" gorm:"foreignKey:InstitutionID"`
}

func (i *Institution) BeforeCreate(tx *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}
