package service

import (
	"errors"
	"go-gin-api-server/internal/model"
	"go-gin-api-server/internal/repository"
	"go-gin-api-server/internal/service"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockPostRepository struct {
	mock.Mock
}

func NewMockPostRepository() *mockPostRepository {
	return &mockPostRepository{}
}

var (
	testPost = &model.Post{
		ID: "1", Title: "Test Post", Content: "Test Content", Author: "Test Author",
	}
	testAllPosts = []model.Post{
		{ID: "1", Title: "Test Post", Content: "Test Content", Author: "Test Author"},
		{ID: "2", Title: "Test Post 2", Content: "Test Content 2", Author: "Test Author 2"},
	}
	updatedPost = &model.Post{
		ID: "1", Title: "Updated Title", Content: "Updated Content", Author: "Updated Author",
	}
)

func setupTestService() (*mockPostRepository, service.PostService) {
	mockRepo := NewMockPostRepository()
	service := service.NewPostService(mockRepo)
	return mockRepo, service
}

// Mock methods

func (m *mockPostRepository) Create(post *model.Post) (*model.Post, error) {
	args := m.Called(post)
	if p := args.Get(0); p != nil {
		return p.(*model.Post), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockPostRepository) FindAll() ([]model.Post, error) {
	args := m.Called()
	if p := args.Get(0); p != nil {
		return p.([]model.Post), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockPostRepository) FindByID(id string) (*model.Post, error) {
	args := m.Called(id)
	if p := args.Get(0); p != nil {
		return p.(*model.Post), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockPostRepository) Update(id string, post *model.Post) (*model.Post, error) {
	args := m.Called(id, post)
	if p := args.Get(0); p != nil {
		return p.(*model.Post), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockPostRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// Testcases

func TestGetPostByID_Success(t *testing.T) {
	mockRepo, service := setupTestService()
	expected := testPost
	mockRepo.On("FindByID", "1").Return(expected, nil)

	// run
	post, err := service.GetByID("1")

	// assert
	assert.NoError(t, err)
	assert.Equal(t, expected, post)
	mockRepo.AssertExpectations(t)
}

func TestGetPostByID_NotFound(t *testing.T) {
	mockRepo, service := setupTestService()
	mockRepo.On("FindByID", "1").Return(nil, repository.ErrNotFound)

	// run
	post, err := service.GetByID("1")

	// assert
	assert.Error(t, err)
	assert.Nil(t, post)
	mockRepo.AssertExpectations(t)
}

func TestGetAllPosts_Success(t *testing.T) {
	mockRepo, service := setupTestService()
	expected := testAllPosts
	mockRepo.On("FindAll").Return(expected, nil)

	// run
	posts, err := service.GetAll()

	// assert
	assert.NoError(t, err)
	assert.Equal(t, expected, posts)
	mockRepo.AssertExpectations(t)
}

func TestGetAllPosts_Error(t *testing.T) {
	mockRepo, service := setupTestService()
	mockRepo.On("FindAll").Return(nil, errors.New("database error"))

	// run
	posts, err := service.GetAll()

	// assert
	assert.Error(t, err)
	assert.Nil(t, posts)
	mockRepo.AssertExpectations(t)
}

func TestCreatePost_Success(t *testing.T) {
	mockRepo, service := setupTestService()
	expected := testPost
	mockRepo.On("Create", mock.Anything).Return(expected, nil)

	// run
	post, err := service.Create(expected)

	// assert
	assert.NoError(t, err)
	assert.Equal(t, expected, post)
	mockRepo.AssertExpectations(t)
}

func TestCreatePost_Error(t *testing.T) {
	mockRepo, service := setupTestService()
	expected := testPost
	mockRepo.On("Create", mock.Anything).Return(nil, errors.New("database error"))

	// run
	post, err := service.Create(expected)

	// assert
	assert.Error(t, err)
	assert.Nil(t, post)
	mockRepo.AssertExpectations(t)
}

func TestUpdatePost_Success(t *testing.T) {
	mockRepo, service := setupTestService()
	expected := updatedPost
	mockRepo.On("Update", "1", mock.Anything).Return(expected, nil)

	// run
	post, err := service.Update("1", expected)

	// assert
	assert.NoError(t, err)
	assert.Equal(t, expected, post)
	mockRepo.AssertExpectations(t)
}

func TestUpdatePost_NotFound(t *testing.T) {
	mockRepo, service := setupTestService()
	expected := updatedPost
	mockRepo.On("Update", "1", mock.Anything).Return(nil, repository.ErrNotFound)

	// run
	post, err := service.Update("1", expected)

	// assert
	assert.Error(t, err)
	assert.Nil(t, post)
	mockRepo.AssertExpectations(t)
}

func TestUpdatePost_Error(t *testing.T) {
	mockRepo, service := setupTestService()
	expected := updatedPost
	mockRepo.On("Update", "1", mock.Anything).Return(nil, errors.New("database error"))

	// run
	post, err := service.Update("1", expected)

	// assert
	assert.Error(t, err)
	assert.Nil(t, post)
	mockRepo.AssertExpectations(t)
}

func TestDeletePost_Success(t *testing.T) {
	mockRepo, service := setupTestService()
	mockRepo.On("Delete", "1").Return(nil)

	// run
	err := service.Delete("1")

	// assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDeletePost_NotFound(t *testing.T) {
	mockRepo, service := setupTestService()
	mockRepo.On("Delete", "1").Return(repository.ErrNotFound)

	// run
	err := service.Delete("1")

	// assert
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDeletePost_Error(t *testing.T) {
	mockRepo, service := setupTestService()
	mockRepo.On("Delete", "1").Return(errors.New("database error"))

	// run
	err := service.Delete("1")

	// assert
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}
