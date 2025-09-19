package service

import (
	"go-gin-api-server/internal/model"

	"github.com/stretchr/testify/mock"
)

type PostServiceMock struct {
	mock.Mock
}

func NewPostServiceMock() *PostServiceMock {
	return &PostServiceMock{}
}

func (m *PostServiceMock) Create(post *model.Post) (*model.Post, error) {
	args := m.Called(post)
	if p := args.Get(0); p != nil {
		postResult, ok := p.(*model.Post)
		if !ok {
			return nil, args.Error(1)
		}
		return postResult, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *PostServiceMock) List(request model.CursorRequest) (*model.CursorResponse[model.PostResponse], error) {
	args := m.Called(request)
	if list := args.Get(0); list != nil {
		listResult, ok := list.(*model.CursorResponse[model.PostResponse])
		if !ok {
			return nil, args.Error(1)
		}
		return listResult, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *PostServiceMock) GetByID(id uint64) (*model.PostResponse, error) {
	args := m.Called(id)
	if p := args.Get(0); p != nil {
		postResult, ok := p.(*model.PostResponse)
		if !ok {
			return nil, args.Error(1)
		}
		return postResult, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *PostServiceMock) Update(id uint64, post *model.Post, currentUserID string) (*model.Post, error) {
	args := m.Called(id, post)
	if p := args.Get(0); p != nil {
		postResult, ok := p.(*model.Post)
		if !ok {
			return nil, args.Error(1)
		}
		return postResult, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *PostServiceMock) Delete(id uint64, currentUserID string) error {
	args := m.Called(id)
	return args.Error(0)
}
