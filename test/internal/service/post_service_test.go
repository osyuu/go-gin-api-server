package service

import (
	"go-gin-api-server/internal/model"
	"go-gin-api-server/internal/service"
	"go-gin-api-server/pkg/apperrors"
	mockRepository "go-gin-api-server/test/mocks/repository"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Helper functions

func setupTestPostService() (*mockRepository.PostRepositoryMock, service.PostService) {
	mockRepo := mockRepository.NewPostRepositoryMock()
	service := service.NewPostService(mockRepo)
	return mockRepo, service
}

func createTestPost(overrides ...map[string]interface{}) *model.Post {
	// Default
	id := uint64(1)
	content := "Test Content"
	createdAt := time.Now()
	updatedAt := time.Now()
	authorID := authorID

	if len(overrides) > 0 {
		override := overrides[0]
		if val, ok := override["id"]; ok {
			id = val.(uint64)
		}

		if val, ok := override["author_id"]; ok {
			authorID = val.(string)
		}

		if val, ok := override["content"]; ok {
			content = val.(string)
		}

		if val, ok := override["created_at"]; ok {
			createdAt = val.(time.Time)
		}

		if val, ok := override["updated_at"]; ok {
			updatedAt = val.(time.Time)
		}
	}

	return &model.Post{
		ID:        id,
		Content:   content,
		AuthorID:  authorID,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

var (
	authorID          = "author-e29b-41d4-a716-446655440000"
	NonExistentPostID = uint64(999)
)

// Testcases

func TestCreatePost(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo, service := setupTestPostService()
		expected := createTestPost()
		repo.On("Create", mock.Anything).Return(expected, nil)

		// run
		created, err := service.Create(expected)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, expected, created)
		repo.AssertExpectations(t)
	})

	t.Run("Content too short", func(t *testing.T) {
		_, service := setupTestPostService()
		post := createTestPost(map[string]interface{}{
			"content": "short", // < 10 characters
		})

		_, err := service.Create(post)
		assert.ErrorIs(t, err, apperrors.ErrPostContentTooShort)
	})

	t.Run("Content too long", func(t *testing.T) {
		_, service := setupTestPostService()
		post := createTestPost(map[string]interface{}{
			"content": strings.Repeat("a", 256), // > 255 characters
		})

		_, err := service.Create(post)
		assert.ErrorIs(t, err, apperrors.ErrPostContentTooLong)
	})

	t.Run("Contains sensitive words", func(t *testing.T) {
		_, service := setupTestPostService()
		post := createTestPost(map[string]interface{}{
			"content": "This contains violence", // contains sensitive words
		})

		_, err := service.Create(post)
		assert.ErrorIs(t, err, apperrors.ErrPostContentSensitiveWords)
	})
}

func TestUpdatePost(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo, service := setupTestPostService()
		created := createTestPost()
		expected := createTestPost(map[string]interface{}{
			"content":    "Updated Content",
			"updated_at": created.CreatedAt.Add(time.Second),
		})
		repo.On("CheckPermission", mock.Anything, mock.Anything).Return(nil)
		repo.On("Update", mock.Anything, mock.Anything).Return(expected, nil)

		// run
		updated, err := service.Update(created.ID, expected, authorID)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, expected.ID, updated.ID)
		assert.Equal(t, expected.Content, updated.Content)
		assert.Equal(t, expected.AuthorID, updated.AuthorID)
		assert.Equal(t, expected.CreatedAt, updated.CreatedAt)
		assert.True(t, updated.UpdatedAt.After(updated.CreatedAt))
		repo.AssertExpectations(t)
	})

	t.Run("ErrorForbidden", func(t *testing.T) {
		repo, service := setupTestPostService()
		created := createTestPost()
		repo.On("CheckPermission", mock.Anything, mock.Anything).Return(apperrors.ErrForbidden)
		repo.On("Update", mock.Anything, mock.Anything).Return(nil, nil)

		// run
		updated, err := service.Update(created.ID, created, authorID)

		// assert
		assert.ErrorIs(t, err, apperrors.ErrForbidden)
		assert.Nil(t, updated)
		repo.AssertNotCalled(t, "Update")
	})

	t.Run("Content too short", func(t *testing.T) {
		repo, service := setupTestPostService()
		created := createTestPost()
		updated := createTestPost(map[string]interface{}{
			"content": "short", // < 10 characters
		})

		repo.On("CheckPermission", mock.Anything, mock.Anything).Return(nil)

		// run
		_, err := service.Update(created.ID, updated, authorID)

		// assert
		assert.ErrorIs(t, err, apperrors.ErrPostContentTooShort)
	})

	t.Run("Content too long", func(t *testing.T) {
		repo, service := setupTestPostService()
		created := createTestPost()
		updated := createTestPost(map[string]interface{}{
			"content": strings.Repeat("a", 256), // > 255 characters
		})

		repo.On("CheckPermission", mock.Anything, mock.Anything).Return(nil)

		// run
		_, err := service.Update(created.ID, updated, authorID)

		// assert
		assert.ErrorIs(t, err, apperrors.ErrPostContentTooLong)
	})

	t.Run("Contains sensitive words", func(t *testing.T) {
		repo, service := setupTestPostService()
		created := createTestPost()
		updated := createTestPost(map[string]interface{}{
			"content": "This contains violence", // contains sensitive words
		})

		repo.On("CheckPermission", mock.Anything, mock.Anything).Return(nil)

		// run
		_, err := service.Update(created.ID, updated, authorID)

		// assert
		assert.ErrorIs(t, err, apperrors.ErrPostContentSensitiveWords)
	})
}

func TestDeletePost(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo, service := setupTestPostService()
		created := createTestPost()
		repo.On("CheckPermission", mock.Anything, mock.Anything).Return(nil)
		repo.On("Delete", mock.Anything).Return(nil)

		// run
		err := service.Delete(created.ID, created.AuthorID)

		// assert
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("ErrorForbidden", func(t *testing.T) {
		repo, service := setupTestPostService()
		created := createTestPost()
		repo.On("CheckPermission", mock.Anything, mock.Anything).Return(apperrors.ErrForbidden)
		repo.On("Delete", mock.Anything).Return(nil)

		// run
		err := service.Delete(created.ID, created.AuthorID)

		// assert
		assert.ErrorIs(t, err, apperrors.ErrForbidden)
		repo.AssertNotCalled(t, "Delete")
	})
}

func TestGetPostByID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo, service := setupTestPostService()
		created := createTestPost()
		repo.On("FindByID", mock.Anything).Return(created, nil)

		// run
		found, err := service.GetByID(created.ID)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, created, found)
		repo.AssertExpectations(t)
	})

	t.Run("ErrorNotFound", func(t *testing.T) {
		repo, service := setupTestPostService()
		repo.On("FindByID", mock.Anything).Return(nil, apperrors.ErrNotFound)

		// run
		found, err := service.GetByID(NonExistentPostID)

		// assert
		assert.ErrorIs(t, err, apperrors.ErrNotFound)
		assert.Nil(t, found)
		repo.AssertExpectations(t)
	})
}

func TestListPosts(t *testing.T) {
	t.Run("Valid limit with no results", func(t *testing.T) {
		repo, service := setupTestPostService()
		expectedOpts := model.PostListOptions{
			Limit:    11, // limit + 1
			AuthorID: nil,
			Cursor:   model.Cursor{ID: "", CreatedAt: time.Time{}},
		}
		repo.On("List", expectedOpts).Return([]model.Post{}, nil)

		request := model.CursorRequest{
			Cursor:   "",
			Limit:    10,
			AuthorID: nil,
		}

		// run
		result, err := service.List(request)

		// assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, []model.Post{}, result.Data)
		assert.Equal(t, "", result.Next)
		assert.Equal(t, false, result.HasMore)
		repo.AssertExpectations(t)
	})

	t.Run("Valid limit with results but no more", func(t *testing.T) {
		repo, service := setupTestPostService()

		// Create test posts (exactly 3 records, less than limit)
		posts := []model.Post{
			*createTestPost(map[string]interface{}{
				"id":         uint64(3),
				"created_at": time.Date(2024, 1, 3, 10, 0, 0, 0, time.UTC),
			}),
			*createTestPost(map[string]interface{}{
				"id":         uint64(2),
				"created_at": time.Date(2024, 1, 2, 9, 0, 0, 0, time.UTC),
			}),
			*createTestPost(map[string]interface{}{
				"id":         uint64(1),
				"created_at": time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC),
			}),
		}

		expectedOpts := model.PostListOptions{
			Limit:    11, // limit + 1
			AuthorID: nil,
			Cursor:   model.Cursor{ID: "", CreatedAt: time.Time{}},
		}
		repo.On("List", expectedOpts).Return(posts, nil)

		request := model.CursorRequest{
			Cursor:   "",
			Limit:    10,
			AuthorID: nil,
		}

		// run
		result, err := service.List(request)

		// assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Data, 3)          // Should return all 3 records
		assert.Equal(t, false, result.HasMore) // Should indicate no more results
		assert.Equal(t, "", result.Next)       // Should have no next cursor
		repo.AssertExpectations(t)
	})

	t.Run("With cursor and has more results", func(t *testing.T) {
		repo, service := setupTestPostService()

		// Create test posts
		posts := []model.Post{
			*createTestPost(map[string]interface{}{
				"id":         uint64(5),
				"created_at": time.Date(2024, 1, 3, 10, 0, 0, 0, time.UTC),
			}),
			*createTestPost(map[string]interface{}{
				"id":         uint64(4),
				"created_at": time.Date(2024, 1, 3, 10, 0, 0, 0, time.UTC),
			}),
			*createTestPost(map[string]interface{}{
				"id":         uint64(3),
				"created_at": time.Date(2024, 1, 3, 10, 0, 0, 0, time.UTC),
			}),
			*createTestPost(map[string]interface{}{
				"id":         uint64(2),
				"created_at": time.Date(2024, 1, 2, 9, 0, 0, 0, time.UTC),
			}),
			*createTestPost(map[string]interface{}{
				"id":         uint64(1),
				"created_at": time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC),
			}),
			// Extra record to indicate has more
			*createTestPost(map[string]interface{}{
				"id":         uint64(0),
				"created_at": time.Date(2024, 1, 1, 7, 0, 0, 0, time.UTC),
			}),
		}

		// Service will request limit + 1 = 6 records
		expectedOpts := model.PostListOptions{
			Limit:    6, // limit + 1
			AuthorID: nil,
			Cursor: model.Cursor{
				ID:        "10",
				CreatedAt: time.Date(2024, 1, 4, 11, 0, 0, 0, time.UTC),
			},
		}
		repo.On("List", expectedOpts).Return(posts, nil)

		// Create cursor string
		cursorStr := model.EncodeCursor(model.Cursor{
			ID:        "10",
			CreatedAt: time.Date(2024, 1, 4, 11, 0, 0, 0, time.UTC),
		})

		request := model.CursorRequest{
			Cursor:   cursorStr,
			Limit:    5,
			AuthorID: nil,
		}

		// run
		result, err := service.List(request)

		// assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Data, 5)         // Should return only 5 records
		assert.Equal(t, true, result.HasMore) // Should indicate more results
		assert.NotEmpty(t, result.Next)       // Should have next cursor

		// Verify next cursor is from the last returned post (ID=1)
		decodedCursor, err := model.DecodeCursor(result.Next)
		assert.NoError(t, err)
		assert.Equal(t, "1", decodedCursor.ID)

		repo.AssertExpectations(t)
	})

	t.Run("Invalid cursor", func(t *testing.T) {
		_, service := setupTestPostService()
		request := model.CursorRequest{
			Cursor:   "invalid",
			Limit:    10,
			AuthorID: nil,
		}
		_, err := service.List(request)
		assert.ErrorIs(t, err, apperrors.ErrValidation)
	})

	t.Run("Limit gets set to default when zero", func(t *testing.T) {
		repo, service := setupTestPostService()
		// Service will request limit + 1 = 11 records (default 10 + 1)
		expectedOpts := model.PostListOptions{
			Limit:    11, // default 10 + 1
			AuthorID: nil,
			Cursor:   model.Cursor{ID: "", CreatedAt: time.Time{}},
		}
		repo.On("List", expectedOpts).Return([]model.Post{}, nil)

		request := model.CursorRequest{
			Cursor:   "",
			Limit:    0, // Will be set to default 10
			AuthorID: nil,
		}

		result, err := service.List(request)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		repo.AssertExpectations(t)
	})

	t.Run("Limit gets capped at 100 when too large", func(t *testing.T) {
		repo, service := setupTestPostService()
		// Service will request limit + 1 = 101 records (capped 100 + 1)
		expectedOpts := model.PostListOptions{
			Limit:    101, // capped 100 + 1
			AuthorID: nil,
			Cursor:   model.Cursor{ID: "", CreatedAt: time.Time{}},
		}
		repo.On("List", expectedOpts).Return([]model.Post{}, nil)

		request := model.CursorRequest{
			Cursor:   "",
			Limit:    150, // Will be capped to 100
			AuthorID: nil,
		}

		result, err := service.List(request)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		repo.AssertExpectations(t)
	})
}
