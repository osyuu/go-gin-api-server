package repository

import (
	"go-gin-api-server/internal/model"
	"go-gin-api-server/pkg/apperrors"
	"sync"
)

type AuthRepository interface {
	CreateCredentials(credentials *model.UserCredentials) (*model.UserCredentials, error)
	FindByUserID(userID string) (*model.UserCredentials, error)
	UpdatePassword(userID string, hashedPassword string) error
	DeleteCredentials(userID string) error
}

type authRepositoryImpl struct {
	mutex       sync.RWMutex
	credentials map[string]*model.UserCredentials
}

func NewAuthRepository() AuthRepository {
	return &authRepositoryImpl{
		credentials: make(map[string]*model.UserCredentials),
	}
}

func (r *authRepositoryImpl) CreateCredentials(credentials *model.UserCredentials) (*model.UserCredentials, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.credentials[credentials.UserID]; exists {
		return nil, apperrors.ErrUserExists
	}

	r.credentials[credentials.UserID] = credentials
	return credentials, nil
}

func (r *authRepositoryImpl) FindByUserID(userID string) (*model.UserCredentials, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	credentials, exists := r.credentials[userID]
	if !exists {
		return nil, apperrors.ErrNotFound
	}

	return credentials, nil
}

func (r *authRepositoryImpl) UpdatePassword(userID string, hashedPassword string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	credentials, exists := r.credentials[userID]
	if !exists {
		return apperrors.ErrNotFound
	}

	credentials.Password = hashedPassword
	return nil
}

func (r *authRepositoryImpl) DeleteCredentials(userID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	_, exists := r.credentials[userID]
	if !exists {
		return apperrors.ErrNotFound
	}

	delete(r.credentials, userID)
	return nil
}
