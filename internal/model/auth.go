package model

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username,omitempty" binding:"omitempty,min=3,max=50,username"`
	Email    string `json:"email,omitempty" binding:"omitempty,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Name      string     `json:"name" binding:"required,min=3"`
	BirthDate *time.Time `json:"birth_date,omitempty"`
	Username  string     `json:"username,omitempty" binding:"omitempty,min=3,max=50,username"`
	Email     string     `json:"email,omitempty" binding:"omitempty,email"`
	Password  string     `json:"password" binding:"required,min=6"`
}

// TokenResponse JWT token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

// Claims JWT claims - store user info in token
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// UserCredentials 用戶認證憑證
type UserCredentials struct {
	UserID   string `json:"user_id"`
	Password string `json:"-"` // 哈希後的密碼，不在JSON中顯示
}
