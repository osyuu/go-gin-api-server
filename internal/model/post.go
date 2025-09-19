package model

import (
	"time"

	"gorm.io/gorm"
)

type Post struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	Content   string    `json:"content" binding:"required,min=10"`
	AuthorID  string    `gorm:"index" json:"author_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// related fields
	Author *User `gorm:"foreignKey:AuthorID" json:"author"`
}

// GORM Hooks
func (p *Post) BeforeCreate(tx *gorm.DB) error {
	now := time.Now().UTC().Truncate(time.Microsecond)
	p.CreatedAt = now
	p.UpdatedAt = now
	return nil
}

func (p *Post) BeforeUpdate(tx *gorm.DB) error {
	p.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)
	return nil
}

// Post DTO
type PostResponse struct {
	Post
	Author *AuthorSummary `json:"author,omitempty"`
}

type AuthorSummary struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Username *string `json:"username,omitempty"`
}

// ListOptions for post list query
type PostListOptions struct {
	AuthorID *string `json:"author_id,omitempty"`
	Limit    int     `json:"limit"`
	Cursor   Cursor  `json:"cursor"`
}
