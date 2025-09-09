package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        string     `json:"id,omitempty"`
	Name      string     `json:"name" binding:"required,min=3"`
	BirthDate *time.Time `json:"birth_date,omitempty"`
	// auth related fields
	Username  string    `json:"username,omitempty" binding:"omitempty,min=3,max=50,username"`
	Email     string    `json:"email,omitempty" binding:"omitempty,email"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateUser creates a new user with the given information
func CreateUser(name, username, email string, birthDate *time.Time) *User {
	now := time.Now().UTC()
	return &User{
		Name:      name,
		Username:  username,
		Email:     email,
		BirthDate: birthDate,
		IsActive:  true, // default to active
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// UpdateUser updates user information
func (u *User) UpdateUser(name, username, email string, birthDate *time.Time) {
	u.Name = name
	u.Username = username
	u.Email = email
	u.BirthDate = birthDate
	u.UpdatedAt = time.Now().UTC()
}

// Deactivate deactivates the user
func (u *User) Deactivate() {
	u.IsActive = false
	u.UpdatedAt = time.Now().UTC()
}

// Activate activates the user
func (u *User) Activate() {
	u.IsActive = true
	u.UpdatedAt = time.Now().UTC()
}

// GORM Hooks
func (u *User) BeforeCreate(tx *gorm.DB) error {
	now := time.Now().UTC()
	if u.CreatedAt.IsZero() {
		u.CreatedAt = now
	}
	if u.UpdatedAt.IsZero() {
		u.UpdatedAt = now
	}
	return nil
}

func (u *User) BeforeUpdate(tx *gorm.DB) error {
	u.UpdatedAt = time.Now().UTC()
	return nil
}
