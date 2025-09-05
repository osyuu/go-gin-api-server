package services

import (
	"go-gin-api-server/internal/model"

	"github.com/stretchr/testify/mock"
)

// AuthServiceMock 是 AuthService 的 mock 實現
type AuthServiceMock struct {
	mock.Mock
}

// NewAuthServiceMock 創建新的 AuthService mock
func NewAuthServiceMock() *AuthServiceMock {
	return &AuthServiceMock{}
}

func (m *AuthServiceMock) Register(req *model.RegisterRequest) (*model.TokenResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.TokenResponse), args.Error(1)
}

func (m *AuthServiceMock) Login(req *model.LoginRequest) (*model.TokenResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.TokenResponse), args.Error(1)
}

func (m *AuthServiceMock) RefreshToken(refreshToken string) (*model.TokenResponse, error) {
	args := m.Called(refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.TokenResponse), args.Error(1)
}

func (m *AuthServiceMock) RefreshAccessToken(refreshToken string) (string, error) {
	args := m.Called(refreshToken)
	return args.String(0), args.Error(1)
}

func (m *AuthServiceMock) ValidateToken(tokenString string) (*model.Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Claims), args.Error(1)
}
