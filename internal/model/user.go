package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        string     `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id,omitempty"`
	Name      string     `gorm:"not null" json:"name" binding:"required,min=3"`
	BirthDate *time.Time `json:"birth_date,omitempty"`
	// auth related fields
	Username  *string   `gorm:"uniqueIndex" json:"username,omitempty" binding:"omitempty,min=3,max=50,username"`
	Email     *string   `gorm:"uniqueIndex" json:"email,omitempty" binding:"omitempty,email"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// related fields
	UserCredentials *UserCredentials `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

// CreateUser creates a new user with the given information
func CreateUser(name string, username *string, email *string, birthDate *time.Time) *User {
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
func (u *User) UpdateUser(name string, username, email *string, birthDate *time.Time) {
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
