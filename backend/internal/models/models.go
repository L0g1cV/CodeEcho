package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	Username     string         `gorm:"uniqueIndex;size:100;not null" json:"username"`
	PasswordHash string         `gorm:"not null" json:"-"`
	LLMConfig    *LLMConfig     `gorm:"type:jsonb" json:"llmConfig,omitempty"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	Notes        []Note         `gorm:"foreignKey:UserID" json:"-"`
}

// LLMConfig stores user's LLM API configuration (BYOK)
type LLMConfig struct {
	Provider string `json:"provider"` // openai, claude, deepseek, custom
	APIKey   string `json:"apiKey"`   // Encrypted in production
	BaseURL  string `json:"baseUrl,omitempty"`
	Model    string `json:"model"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// NoteType represents the type of note
type NoteType string

const (
	NoteTypeNotebook NoteType = "notebook"
	NoteTypeFolder   NoteType = "folder"
	NoteTypeNote     NoteType = "note"
)

// Note represents a unified entity that can be a notebook, folder, or note
// - notebook: ParentID is null, acts as a project/collection
// - folder: ParentID points to notebook or parent folder, organizes notes
// - note: ParentID points to notebook or folder, contains actual content
type Note struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	UserID       uuid.UUID      `gorm:"type:uuid;not null;index" json:"userId"`
	ParentID     *uuid.UUID     `gorm:"type:uuid;index" json:"parentId,omitempty"`
	Type         NoteType       `gorm:"type:varchar(20);not null;index;default:'note'" json:"type"`
	Title        string         `gorm:"size:500;not null" json:"title"`
	Content      string         `gorm:"type:text" json:"content"`              // Markdown content (for all types)
	ProblemURL   string         `gorm:"size:1000" json:"problemUrl,omitempty"` // Only for type="note"
	SolutionCode string         `gorm:"type:text" json:"solutionCode"`         // Only for type="note"
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	Children     []Note         `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}

func (n *Note) BeforeCreate(tx *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return nil
}

// IsNotebook returns true if this note is a notebook (root level)
func (n *Note) IsNotebook() bool {
	return n.Type == NoteTypeNotebook
}

// IsFolder returns true if this note is a folder
func (n *Note) IsFolder() bool {
	return n.Type == NoteTypeFolder
}

// IsNote returns true if this note is a regular note
func (n *Note) IsNote() bool {
	return n.Type == NoteTypeNote
}
