package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReportStatus string

const (
	ReportPending    ReportStatus = "pending"
	ReportVerified   ReportStatus = "verified"
	ReportInProgress ReportStatus = "in_progress"
	ReportResolved   ReportStatus = "resolved"
	ReportRejected   ReportStatus = "rejected"
)

type ReportSeverity string

const (
	SeverityLow      ReportSeverity = "low"
	SeverityMedium   ReportSeverity = "medium"
	SeverityHigh     ReportSeverity = "high"
	SeverityCritical ReportSeverity = "critical"
)

type Report struct {
	ID            uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	Title         string         `json:"title" gorm:"not null"`
	Description   string         `json:"description" gorm:"type:text;not null"`
	CategoryID    uuid.UUID      `json:"category_id" gorm:"type:uuid;not null"`
	Category      Category       `json:"category" gorm:"foreignKey:CategoryID"`
	ReporterID    uuid.UUID      `json:"reporter_id" gorm:"type:uuid"`
	Reporter      *User          `json:"reporter,omitempty" gorm:"foreignKey:ReporterID"`
	IsAnonymous   bool           `json:"is_anonymous" gorm:"default:false"`
	Status        ReportStatus   `json:"status" gorm:"type:varchar(20);default:'pending'"`
	Severity      ReportSeverity `json:"severity" gorm:"type:varchar(10);default:'medium'"`
	Latitude      float64        `json:"latitude"`
	Longitude     float64        `json:"longitude"`
	Province      string         `json:"province" gorm:"not null"`
	City          string         `json:"city"`
	Address       string         `json:"address"`
	OccurredAt    time.Time      `json:"occurred_at"`
	VictimsCount  int            `json:"victims_count" gorm:"default:0"`
	Tags          string         `json:"tags"`
	ModeratorNote string         `json:"moderator_note"`
	Evidences     []ReportEvidence `json:"evidences,omitempty" gorm:"foreignKey:ReportID"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

func (r *Report) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

type ReportEvidence struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	ReportID  uuid.UUID `json:"report_id" gorm:"type:uuid;not null"`
	FileURL   string    `json:"file_url" gorm:"not null"`
	FileType  string    `json:"file_type"`
	Caption   string    `json:"caption"`
	CreatedAt time.Time `json:"created_at"`
}

func (e *ReportEvidence) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}
