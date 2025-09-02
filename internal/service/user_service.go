package service

import (
	"blog_server/internal/model"
	"blog_server/internal/repository"
	"errors"
)

type UserService interface {
	GetUserById(id string) (*model.User, error)
	CreateUser(user *model.User) error
}

type userServiceImpl struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userServiceImpl{repo: repo}
}

func (s *userServiceImpl) GetUserById(id string) (*model.User, error) {
	user, err := s.repo.GetUserById(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (s *userServiceImpl) CreateUser(user *model.User) error {
	err := s.repo.CreateUser(user)
	if err != nil {
		return err
	}
	return nil
}
