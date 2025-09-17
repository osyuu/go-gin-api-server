package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        string     `gorm:"primaryKey;default:gen_random_uuid()" json:"id,omitempty"`
	Name      string     `json:"name"`
	BirthDate *time.Time `json:"birth_date,omitempty"`
	// auth related fields
	Username  *string   `gorm:"uniqueIndex" json:"username,omitempty"`
	Email     *string   `gorm:"uniqueIndex" json:"email,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// related fields
	UserCredentials *UserCredentials `gorm:"foreignKey:UserID" json:"-"`
}

// CreateUser creates a new user with the given information
func CreateUser(name string, username *string, email *string, birthDate *time.Time) *User {
	return &User{
		Name:      name,
		Username:  username,
		Email:     email,
		BirthDate: birthDate,
		IsActive:  true, // default to active
	}
}

// GORM Hooks
func (u *User) BeforeCreate(tx *gorm.DB) error {
	now := time.Now().UTC().Truncate(time.Microsecond)
	u.CreatedAt = now
	u.UpdatedAt = now
	return nil
}

func (u *User) BeforeUpdate(tx *gorm.DB) error {
	u.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)
	return nil
}

// User external structures
type UpdateUserProfileRequest struct {
	Name      string     `json:"name,omitempty" binding:"omitempty,min=3"`
	BirthDate *time.Time `json:"birth_date,omitempty"`
}

type UserProfile struct {
	Name      string     `json:"name"`
	Username  *string    `json:"username,omitempty"`
	BirthDate *time.Time `json:"birth_date,omitempty"`
}
