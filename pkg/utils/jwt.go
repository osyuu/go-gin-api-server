package utils

import (
	"errors"
	"go-gin-api-server/internal/model"
	"go-gin-api-server/pkg/apperrors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	secretKey     string
	tokenDuration time.Duration
}

func NewJWTManager(secretKey string, tokenDuration time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:     secretKey,
		tokenDuration: tokenDuration,
	}
}

func (j *JWTManager) GenerateToken(user *model.User) (*model.TokenResponse, error) {
	// generate access token
	tokenString, err := j.GenerateAccessToken(user)
	if err != nil {
		return nil, err
	}

	// generate refresh token
	now := time.Now().UTC()
	refreshExpiresAt := now.Add(j.tokenDuration * 24 * 7) // 7 days
	refreshClaims := &model.Claims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "go-gin-api-server",
			Subject:   user.ID,
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(j.secretKey))
	if err != nil {
		return nil, err
	}

	return &model.TokenResponse{
		AccessToken:  tokenString,
		RefreshToken: refreshTokenString,
		TokenType:    "Bearer",
		ExpiresIn:    int64(j.tokenDuration.Seconds()),
	}, nil
}

// GenerateAccessTokenOnly
func (j *JWTManager) GenerateAccessToken(user *model.User) (string, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(j.tokenDuration)

	claims := &model.Claims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "go-gin-api-server",
			Subject:   user.ID,
		},
	}

	// generate access token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GetTokenDuration 獲取 token 有效期
func (j *JWTManager) GetTokenDuration() time.Duration {
	return j.tokenDuration
}

// ValidateToken 驗證JWT token
func (j *JWTManager) ValidateToken(tokenString string) (*model.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &model.Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 驗證簽名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(j.secretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, apperrors.ErrExpiredToken
		}
		return nil, apperrors.ErrInvalidToken
	}

	claims, ok := token.Claims.(*model.Claims)
	if !ok || !token.Valid {
		return nil, apperrors.ErrInvalidToken
	}

	return claims, nil
}
