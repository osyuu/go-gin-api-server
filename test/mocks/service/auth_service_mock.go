package service

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
	if resp := args.Get(0); resp != nil {
		response, ok := resp.(*model.TokenResponse)
		if !ok {
			return nil, args.Error(1)
		}
		err := args.Error(1)
		return response, err
	}
	err := args.Error(1)
	return nil, err
}

func (m *AuthServiceMock) Login(req *model.LoginRequest) (*model.TokenResponse, error) {
	args := m.Called(req)
	if resp := args.Get(0); resp != nil {
		response, ok := resp.(*model.TokenResponse)
		if !ok {
			return nil, args.Error(1)
		}
		err := args.Error(1)
		return response, err
	}
	err := args.Error(1)
	return nil, err
}

func (m *AuthServiceMock) RefreshToken(refreshToken string) (*model.TokenResponse, error) {
	args := m.Called(refreshToken)
	if resp := args.Get(0); resp != nil {
		response, ok := resp.(*model.TokenResponse)
		if !ok {
			return nil, args.Error(1)
		}
		err := args.Error(1)
		return response, err
	}
	err := args.Error(1)
	return nil, err
}

func (m *AuthServiceMock) RefreshAccessToken(refreshToken string) (string, error) {
	args := m.Called(refreshToken)
	token := args.String(0)
	err := args.Error(1)
	return token, err
}

func (m *AuthServiceMock) ValidateToken(tokenString string) (*model.Claims, error) {
	args := m.Called(tokenString)
	if claims := args.Get(0); claims != nil {
		claimsResult, ok := claims.(*model.Claims)
		if !ok {
			return nil, args.Error(1)
		}
		err := args.Error(1)
		return claimsResult, err
	}
	err := args.Error(1)
	return nil, err
}
