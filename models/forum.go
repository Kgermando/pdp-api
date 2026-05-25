package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ForumCategory string

const (
	ForumCatGovernance     ForumCategory = "governance"
	ForumCatEducation      ForumCategory = "education"
	ForumCatHealth         ForumCategory = "health"
	ForumCatAgriculture    ForumCategory = "agriculture"
	ForumCatInfrastructure ForumCategory = "infrastructure"
	ForumCatSecurity       ForumCategory = "security"
	ForumCatEconomy        ForumCategory = "economy"
	ForumCatEnvironment    ForumCategory = "environment"
	ForumCatCulture        ForumCategory = "culture"
	ForumCatOther          ForumCategory = "other"
)

// Forum represents a public discussion salon
type Forum struct {
	ID             uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	Title          string         `json:"title" gorm:"not null"`
	Description    string         `json:"description" gorm:"type:text"`
	Category       ForumCategory  `json:"category" gorm:"type:varchar(30);not null"`
	AuthorID       uuid.UUID      `json:"author_id" gorm:"type:uuid;not null"`
	Author         User           `json:"author" gorm:"foreignKey:AuthorID"`
	Province       string         `json:"province"`
	Tags           string         `json:"tags"`
	IsAnonymous    bool           `json:"is_anonymous" gorm:"default:false"`
	IsPinned       bool           `json:"is_pinned" gorm:"default:false"`
	IsLocked       bool           `json:"is_locked" gorm:"default:false"`
	MessageCount   int            `json:"message_count" gorm:"default:0"`
	ViewsCount     int            `json:"views_count" gorm:"default:0"`
	LikesCount     int            `json:"likes_count" gorm:"default:0"`
	LastActivityAt *time.Time     `json:"last_activity_at"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
	Messages       []ForumMessage `json:"messages,omitempty" gorm:"foreignKey:ForumID"`
}

func (f *Forum) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return nil
}

// ForumMessage represents a message inside a forum
type ForumMessage struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	ForumID     uuid.UUID      `json:"forum_id" gorm:"type:uuid;not null"`
	ParentID    *uuid.UUID     `json:"parent_id" gorm:"type:uuid"` // nil = top-level, set = reply
	AuthorID    uuid.UUID      `json:"author_id" gorm:"type:uuid;not null"`
	Author      User           `json:"author" gorm:"foreignKey:AuthorID"`
	Content     string         `json:"content" gorm:"type:text;not null"`
	IsAnonymous bool           `json:"is_anonymous" gorm:"default:false"`
	IsOfficial  bool           `json:"is_official" gorm:"default:false"` // institution or admin response
	LikesCount  int            `json:"likes_count" gorm:"default:0"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	Replies     []ForumMessage `json:"replies,omitempty" gorm:"foreignKey:ParentID"`
}

func (m *ForumMessage) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// ForumLike tracks likes on forums and messages
type ForumLike struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey"`
	ForumID   *uuid.UUID `json:"forum_id" gorm:"type:uuid"`
	MessageID *uuid.UUID `json:"message_id" gorm:"type:uuid"`
	UserID    uuid.UUID  `json:"user_id" gorm:"type:uuid;not null"`
	CreatedAt time.Time  `json:"created_at"`
}

func (l *ForumLike) BeforeCreate(tx *gorm.DB) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return nil
}
