package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProposalStatus string

const (
	ProposalDraft      ProposalStatus = "draft"
	ProposalPending    ProposalStatus = "pending"
	ProposalPublished  ProposalStatus = "published"
	ProposalApproved   ProposalStatus = "approved"
	ProposalRejected   ProposalStatus = "rejected"
	ProposalSubmitted  ProposalStatus = "submitted_to_parliament"
)

type VoteType string

const (
	VoteFor     VoteType = "for"
	VoteAgainst VoteType = "against"
	VoteAbstain VoteType = "abstain"
)

type Proposal struct {
	ID             uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	Title          string         `json:"title" gorm:"not null"`
	Summary        string         `json:"summary" gorm:"not null"`
	Content        string         `json:"content" gorm:"type:text;not null"`
	CategoryID     uuid.UUID      `json:"category_id" gorm:"type:uuid;not null"`
	Category       Category       `json:"category" gorm:"foreignKey:CategoryID"`
	AuthorID       uuid.UUID      `json:"author_id" gorm:"type:uuid;not null"`
	Author         User           `json:"author" gorm:"foreignKey:AuthorID"`
	Status         ProposalStatus `json:"status" gorm:"type:varchar(30);default:'draft'"`
	Province       string         `json:"province"`
	Tags           string         `json:"tags"`
	VotesFor       int            `json:"votes_for" gorm:"default:0"`
	VotesAgainst   int            `json:"votes_against" gorm:"default:0"`
	VotesAbstain   int            `json:"votes_abstain" gorm:"default:0"`
	CommentsCount  int            `json:"comments_count" gorm:"default:0"`
	ViewsCount     int            `json:"views_count" gorm:"default:0"`
	IsAnonymous    bool           `json:"is_anonymous" gorm:"default:false"`
	ModeratorNote  string         `json:"moderator_note"`
	SubmittedAt    *time.Time     `json:"submitted_at"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
	Votes          []ProposalVote    `json:"votes,omitempty" gorm:"foreignKey:ProposalID"`
	Comments       []ProposalComment `json:"comments,omitempty" gorm:"foreignKey:ProposalID"`
}

func (p *Proposal) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

type ProposalVote struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	ProposalID uuid.UUID `json:"proposal_id" gorm:"type:uuid;not null;uniqueIndex:idx_proposal_user_vote"`
	UserID     uuid.UUID `json:"user_id" gorm:"type:uuid;not null;uniqueIndex:idx_proposal_user_vote"`
	VoteType   VoteType  `json:"vote_type" gorm:"type:varchar(10);not null"`
	CreatedAt  time.Time `json:"created_at"`
}

func (v *ProposalVote) BeforeCreate(tx *gorm.DB) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return nil
}

type ProposalComment struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	ProposalID uuid.UUID      `json:"proposal_id" gorm:"type:uuid;not null"`
	UserID     uuid.UUID      `json:"user_id" gorm:"type:uuid;not null"`
	User       User           `json:"user" gorm:"foreignKey:UserID"`
	Content    string         `json:"content" gorm:"type:text;not null"`
	IsHidden   bool           `json:"is_hidden" gorm:"default:false"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}

func (c *ProposalComment) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
