package service

import (
	"go-gin-api-server/internal/model"
	"go-gin-api-server/internal/repository"
	"go-gin-api-server/pkg/apperrors"
	"go-gin-api-server/pkg/utils"
	"slices"
	"time"

	"github.com/bytedance/gopkg/util/logger"
)

type AuthService interface {
	Register(req *model.RegisterRequest) (*model.TokenResponse, error)
	Login(req *model.LoginRequest) (*model.TokenResponse, error)
	RefreshToken(refreshToken string) (*model.TokenResponse, error)
	RefreshAccessToken(refreshToken string) (string, error)
	ValidateToken(tokenString string) (*model.Claims, error)
}

type authServiceImpl struct {
	userRepo repository.UserRepository
	authRepo repository.AuthRepository
	jwtMgr   *utils.JWTManager
}

func NewAuthService(userRepo repository.UserRepository, authRepo repository.AuthRepository, jwtMgr *utils.JWTManager) AuthService {
	return &authServiceImpl{
		userRepo: userRepo,
		authRepo: authRepo,
		jwtMgr:   jwtMgr,
	}
}

func (s *authServiceImpl) Register(req *model.RegisterRequest) (*model.TokenResponse, error) {
	// 1. validate
	if req.BirthDate != nil {
		if s.isUnder13(*req.BirthDate) {
			return nil, apperrors.ErrUserUnderAge
		}
	}

	// Check reserved username only if username is provided
	if req.Username != "" && s.isReservedUsername(req.Username) {
		return nil, apperrors.ErrValidation
	}

	// 2. create user
	var username, email *string
	if req.Username != "" {
		username = &req.Username
	}
	if req.Email != "" {
		email = &req.Email
	}

	user := model.CreateUser(req.Name, username, email, req.BirthDate)

	user, err := s.userRepo.Create(user)
	if err != nil {
		return nil, err
	}

	// 3. hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		// if password hash failed, delete the created user
		if deleteErr := s.userRepo.Delete(user.ID); deleteErr != nil {
			logger.Errorf("Failed to delete user: %v", deleteErr)
		}
		return nil, err
	}

	// 4. create credentials
	credentials := &model.UserCredentials{
		UserID:   user.ID,
		Password: hashedPassword,
	}

	_, err = s.authRepo.CreateCredentials(credentials)
	if err != nil {
		// if credentials creation failed, delete the created user
		if deleteErr := s.userRepo.Delete(user.ID); deleteErr != nil {
			logger.Errorf("Failed to delete user: %v", deleteErr)
		}
		return nil, err
	}

	// 5. generate JWT token
	return s.jwtMgr.GenerateToken(user)
}

func (s *authServiceImpl) Login(req *model.LoginRequest) (*model.TokenResponse, error) {
	// 1. find user
	var user *model.User
	var err error

	if req.Username != "" {
		user, err = s.userRepo.FindByUsername(req.Username)
	} else {
		user, err = s.userRepo.FindByEmail(req.Email)
	}

	if err != nil {
		return nil, apperrors.ErrUnauthorized
	}

	// 2. get credentials
	credentials, err := s.authRepo.FindByUserID(user.ID)
	if err != nil {
		return nil, apperrors.ErrUnauthorized
	}

	// 3. validate password
	err = utils.CheckPassword(credentials.Password, req.Password)
	if err != nil {
		return nil, apperrors.ErrUnauthorized
	}

	// 4. check if user is active
	if !user.IsActive {
		return nil, apperrors.ErrForbidden
	}

	// 5. generate JWT token
	return s.jwtMgr.GenerateToken(user)
}

func (s *authServiceImpl) RefreshToken(refreshToken string) (*model.TokenResponse, error) {
	claims, err := s.jwtMgr.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// check if user is still exists and active
	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, apperrors.ErrUnauthorized
	}
	if !user.IsActive {
		return nil, apperrors.ErrForbidden
	}

	return s.jwtMgr.GenerateToken(user)
}

func (s *authServiceImpl) RefreshAccessToken(refreshToken string) (string, error) {
	claims, err := s.jwtMgr.ValidateToken(refreshToken)
	if err != nil {
		return "", err
	}

	// check if user is still exists and active
	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return "", apperrors.ErrUnauthorized
	}
	if !user.IsActive {
		return "", apperrors.ErrForbidden
	}

	// 只生成新的 Access Token，不生成新的 Refresh Token
	return s.jwtMgr.GenerateAccessToken(user)
}

func (s *authServiceImpl) ValidateToken(tokenString string) (*model.Claims, error) {
	return s.jwtMgr.ValidateToken(tokenString)
}

// 業務邏輯驗證輔助方法

// isUnder13 檢查用戶是否未滿13歲
func (s *authServiceImpl) isUnder13(birthDate time.Time) bool {
	now := time.Now().UTC().Truncate(time.Microsecond)
	age := now.Year() - birthDate.Year()

	// 如果還沒到生日，年齡減1
	if now.YearDay() < birthDate.YearDay() {
		age--
	}

	return age < 13
}

// isReservedUsername 檢查用戶名是否為保留字
func (s *authServiceImpl) isReservedUsername(username string) bool {
	reservedUsernames := []string{
		"admin", "administrator", "root", "system", "api",
		"www", "mail", "ftp", "support", "help",
		"test", "demo", "guest", "user", "default",
	}

	return slices.Contains(reservedUsernames, username)
}
