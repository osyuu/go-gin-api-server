package repository

import (
	"go-gin-api-server/internal/model"

	"github.com/stretchr/testify/mock"
)

type PostRepositoryMock struct {
	mock.Mock
}

func NewPostRepositoryMock() *PostRepositoryMock {
	return &PostRepositoryMock{}
}

// Mock methods

func (m *PostRepositoryMock) Create(post *model.Post) (*model.Post, error) {
	args := m.Called(post)
	if postResult := args.Get(0); postResult != nil {
		post, ok := postResult.(*model.Post)
		if !ok {
			return nil, args.Error(1)
		}
		err := args.Error(1)
		return post, err
	}
	err := args.Error(1)
	return nil, err
}

func (m *PostRepositoryMock) List(opts model.PostListOptions) ([]model.Post, error) {
	args := m.Called(opts)
	if posts := args.Get(0); posts != nil {
		postResult, ok := posts.([]model.Post)
		if !ok {
			return nil, args.Error(1)
		}
		err := args.Error(1)
		return postResult, err
	}
	return nil, args.Error(1)
}

func (m *PostRepositoryMock) FindByID(id uint64) (*model.Post, error) {
	args := m.Called(id)
	if post := args.Get(0); post != nil {
		postResult, ok := post.(*model.Post)
		if !ok {
			return nil, args.Error(1)
		}
		err := args.Error(1)
		return postResult, err
	}
	err := args.Error(1)
	return nil, err
}

func (m *PostRepositoryMock) Update(id uint64, post *model.Post) (*model.Post, error) {
	args := m.Called(id, post)
	if u := args.Get(0); u != nil {
		postResult, ok := u.(*model.Post)
		if !ok {
			return nil, args.Error(1)
		}
		err := args.Error(1)
		return postResult, err
	}
	err := args.Error(1)
	return nil, err
}

func (m *PostRepositoryMock) Delete(id uint64) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *PostRepositoryMock) CheckPermission(id uint64, userID string) error {
	args := m.Called(id, userID)
	return args.Error(0)
}
