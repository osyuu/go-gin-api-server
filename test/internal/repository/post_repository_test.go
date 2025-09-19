package repository

import (
	"go-gin-api-server/internal/model"
	"go-gin-api-server/internal/repository"
	"go-gin-api-server/pkg/apperrors"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func firstCreateTestUser(t *testing.T, tx *gorm.DB, overrides map[string]interface{}) *model.User {
	// First create a user
	userRepo := repository.NewUserRepositoryWithDB(tx)
	user := createTestUser(overrides)
	createdUser, err := userRepo.Create(user)
	assert.NoError(t, err)
	return createdUser
}

func createTestPost(authorID string, overrides ...map[string]interface{}) *model.Post {
	// Default
	content := "Test Content"

	if len(overrides) > 0 {
		override := overrides[0]
		if val, ok := override["content"]; ok {
			content = val.(string)
		}
	}

	return &model.Post{
		Content:  content,
		AuthorID: authorID,
	}
}

// TestCases

func TestCreatePostAndFindByID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		tx := setup()
		defer teardown(tx)

		// First create a user
		createdUser := firstCreateTestUser(t, tx, nil)

		repo := repository.NewPostRepositoryWithDB(tx)
		post := createTestPost(createdUser.ID)

		// run
		created, err := repo.Create(post)
		assert.NoError(t, err)
		found, err := repo.FindByID(created.ID)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, created.ID, found.ID)
		assert.Equal(t, created.Content, found.Content)
		assert.Equal(t, created.AuthorID, found.AuthorID)
		assert.Equal(t, created.CreatedAt.UTC(), found.CreatedAt.UTC())
		assert.Equal(t, created.UpdatedAt.UTC(), found.UpdatedAt.UTC())
		assert.NotNil(t, found.Author, "found.Author should not be nil")
		assert.Equal(t, found.Author.ID, createdUser.ID)
		assert.Equal(t, found.Author.Name, createdUser.Name)
		assert.Equal(t, *found.Author.Username, *createdUser.Username)
		assert.Nil(t, found.Author.Email)
	})

	t.Run("NotFound", func(t *testing.T) {
		tx := setup()
		defer teardown(tx)

		repo := repository.NewPostRepositoryWithDB(tx)

		found, err := repo.FindByID(NonExistentPostID)

		assert.ErrorIs(t, err, apperrors.ErrNotFound)
		assert.Nil(t, found)
	})

	t.Run("Edge Cases", func(t *testing.T) {
		t.Run("Database TEXT Field Limit", func(t *testing.T) {
			tx := setup()
			defer teardown(tx)

			user := firstCreateTestUser(t, tx, nil)
			repo := repository.NewPostRepositoryWithDB(tx)

			// Test the actual limit of the TEXT field in the database
			longContent := strings.Repeat("A", 100000) // 100KB
			post := createTestPost(user.ID,
				map[string]interface{}{
					"content": longContent,
				})

			_, err := repo.Create(post)
			// If the database has TEXT limit, this may fail
			// This depends on the PostgreSQL TEXT field limit
			assert.NoError(t, err)
		})

		t.Run("Foreign Key Constraint", func(t *testing.T) {
			tx := setup()
			defer teardown(tx)

			repo := repository.NewPostRepositoryWithDB(tx)
			post := createTestPost(NonExistentUserID)

			_, err := repo.Create(post)
			// Should be a foreign key constraint error
			assert.Error(t, err)
			// Check if it's a foreign key constraint error
			assert.Contains(t, err.Error(), "foreign key")
		})

	})
}

func TestUpdatePost(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		tx := setup()
		defer teardown(tx)

		// First create a user
		createdUser := firstCreateTestUser(t, tx, nil)

		repo := repository.NewPostRepositoryWithDB(tx)
		post := createTestPost(createdUser.ID)

		// run
		created, err := repo.Create(post)
		assert.NoError(t, err)
		updated := &model.Post{
			Content: "Updated Content",
		}
		found, err := repo.Update(created.ID, updated)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, created.ID, found.ID)
		assert.Equal(t, updated.Content, found.Content)
		assert.Equal(t, created.AuthorID, found.AuthorID)
		assert.Equal(t, created.CreatedAt.UTC(), found.CreatedAt.UTC())
		assert.True(t, found.UpdatedAt.After(found.CreatedAt))
	})

	t.Run("NotFound", func(t *testing.T) {
		tx := setup()
		defer teardown(tx)

		repo := repository.NewPostRepositoryWithDB(tx)

		updated := &model.Post{
			Content: "Updated Content",
		}
		found, err := repo.Update(NonExistentPostID, updated)

		assert.ErrorIs(t, err, apperrors.ErrNotFound)
		assert.Nil(t, found)
	})

	t.Run("Edge Cases", func(t *testing.T) {
		t.Run("Update with Very Long Content", func(t *testing.T) {
			tx := setup()
			defer teardown(tx)

			user := firstCreateTestUser(t, tx, nil)
			repo := repository.NewPostRepositoryWithDB(tx)
			post := createTestPost(user.ID)

			created, err := repo.Create(post)
			assert.NoError(t, err)

			// Update to very long content
			longContent := strings.Repeat("B", 10000)
			updated := &model.Post{
				Content: longContent,
			}
			found, err := repo.Update(created.ID, updated)
			assert.NoError(t, err)
			assert.Equal(t, longContent, found.Content)
		})

	})
}

func TestDeletePost(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		tx := setup()
		defer teardown(tx)

		// First create a user
		createdUser := firstCreateTestUser(t, tx, nil)

		repo := repository.NewPostRepositoryWithDB(tx)
		post := &model.Post{
			Content:  "Test Content",
			AuthorID: createdUser.ID,
		}

		// run
		created, err := repo.Create(post)
		assert.NoError(t, err)
		repo.Delete(created.ID)
		found, err := repo.FindByID(created.ID)

		// assert
		assert.ErrorIs(t, err, apperrors.ErrNotFound)
		assert.Nil(t, found)
	})

	t.Run("NotFound", func(t *testing.T) {
		tx := setup()
		defer teardown(tx)

		repo := repository.NewPostRepositoryWithDB(tx)

		err := repo.Delete(NonExistentPostID)

		assert.ErrorIs(t, err, apperrors.ErrNotFound)
	})

}

func TestListWithCursor(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		tx := setup()
		defer teardown(tx)

		// First create a user
		user := firstCreateTestUser(t, tx, nil)

		// Create 3 posts
		repo := repository.NewPostRepositoryWithDB(tx)
		posts := []*model.Post{
			{Content: "Post 1", AuthorID: user.ID},
			{Content: "Post 2", AuthorID: user.ID},
			{Content: "Post 3", AuthorID: user.ID},
		}

		var createdPosts []model.Post
		for _, post := range posts {
			created, err := repo.Create(post)
			assert.NoError(t, err)
			createdPosts = append(createdPosts, *created)
			time.Sleep(1 * time.Millisecond)
		}

		// Test first page (limit=2)
		firstCursor := model.Cursor{
			CreatedAt: time.Now().Add(1 * time.Hour), // Future time
		}
		firstOptions := model.PostListOptions{
			Cursor: firstCursor,
			Limit:  2,
		}
		firstPage, err := repo.List(firstOptions)
		assert.NoError(t, err)
		assert.Len(t, firstPage, 2)

		// Verify sorting (latest first)
		assert.Equal(t, createdPosts[2].Content, firstPage[0].Content)
		assert.Equal(t, createdPosts[1].Content, firstPage[1].Content)

		// Test second page
		secondCursor := model.Cursor{
			ID:        strconv.FormatUint(firstPage[1].ID, 10),
			CreatedAt: firstPage[1].CreatedAt,
		}
		secondOptions := model.PostListOptions{
			Cursor: secondCursor,
			Limit:  2,
		}
		secondPage, err := repo.List(secondOptions)
		assert.NoError(t, err)
		assert.Len(t, secondPage, 1) // Only one post left
		assert.Equal(t, createdPosts[0].Content, secondPage[0].Content)
	})

	t.Run("Empty List", func(t *testing.T) {
		tx := setup()
		defer teardown(tx)

		repo := repository.NewPostRepositoryWithDB(tx)
		cursor := model.Cursor{
			CreatedAt: time.Now(),
		}
		options := model.PostListOptions{
			Cursor: cursor,
			Limit:  10,
		}
		posts, err := repo.List(options)
		assert.NoError(t, err)
		assert.Len(t, posts, 0)
	})

	t.Run("Edge Cases", func(t *testing.T) {
		t.Run("Very Large Limit", func(t *testing.T) {
			tx := setup()
			defer teardown(tx)

			user := firstCreateTestUser(t, tx, nil)
			repo := repository.NewPostRepositoryWithDB(tx)

			post := createTestPost(user.ID)
			_, err := repo.Create(post)
			assert.NoError(t, err)

			// Test very large limit
			cursor := model.Cursor{
				CreatedAt: time.Now().Add(1 * time.Hour),
			}
			options := model.PostListOptions{
				Cursor: cursor,
				Limit:  1000, // limit = 1000
			}
			posts, err := repo.List(options)
			assert.NoError(t, err)
			assert.Len(t, posts, 1)
		})

		t.Run("Zero Limit", func(t *testing.T) {
			tx := setup()
			defer teardown(tx)

			user := firstCreateTestUser(t, tx, nil)
			repo := repository.NewPostRepositoryWithDB(tx)

			post := createTestPost(user.ID)
			_, err := repo.Create(post)
			assert.NoError(t, err)

			// Test zero limit
			cursor := model.Cursor{
				CreatedAt: time.Now().Add(1 * time.Hour),
			}
			options := model.PostListOptions{
				Cursor: cursor,
				Limit:  0, // limit = 0
			}
			posts, err := repo.List(options)
			assert.NoError(t, err)
			assert.Len(t, posts, 0)
		})

		t.Run("Negative Limit", func(t *testing.T) {
			tx := setup()
			defer teardown(tx)

			user := firstCreateTestUser(t, tx, nil)
			repo := repository.NewPostRepositoryWithDB(tx)

			post := createTestPost(user.ID)
			_, err := repo.Create(post)
			assert.NoError(t, err)

			cursor := model.Cursor{
				CreatedAt: time.Now().Add(1 * time.Hour),
			}
			options := model.PostListOptions{
				Cursor: cursor,
				Limit:  -1, // limit = -1
			}

			posts, err := repo.List(options)
			assert.ErrorIs(t, err, apperrors.ErrValidation)
			assert.Nil(t, posts)
		})

		t.Run("Invalid Cursor ID", func(t *testing.T) {
			tx := setup()
			defer teardown(tx)

			repo := repository.NewPostRepositoryWithDB(tx)
			cursor := model.Cursor{
				ID:        InvalidCursorID,
				CreatedAt: time.Now(),
			}

			options := model.PostListOptions{
				Cursor: cursor,
				Limit:  10,
			}
			_, err := repo.List(options)
			assert.ErrorIs(t, err, apperrors.ErrValidation)
		})
	})

	t.Run("Filter by AuthorID", func(t *testing.T) {
		tx := setup()
		defer teardown(tx)

		// Create two users with different usernames
		user1 := firstCreateTestUser(t, tx, nil)
		user2 := firstCreateTestUser(t, tx, map[string]interface{}{
			"username": "testuser2",
			"email":    "testuser2@test.com",
		})

		// Create posts for both users
		repo := repository.NewPostRepositoryWithDB(tx)
		posts := []*model.Post{
			{Content: "User1 Post 1", AuthorID: user1.ID},
			{Content: "User1 Post 2", AuthorID: user1.ID},
			{Content: "User2 Post 1", AuthorID: user2.ID},
			{Content: "User2 Post 2", AuthorID: user2.ID},
		}

		for _, post := range posts {
			_, err := repo.Create(post)
			assert.NoError(t, err)
			time.Sleep(1 * time.Millisecond)
		}

		// Test filtering by user1
		cursor := model.Cursor{
			CreatedAt: time.Now().Add(1 * time.Hour), // Future time
		}
		options := model.PostListOptions{
			Cursor:   cursor,
			Limit:    10,
			AuthorID: &user1.ID,
		}
		user1Posts, err := repo.List(options)
		assert.NoError(t, err)
		assert.Len(t, user1Posts, 2)

		// Verify all posts belong to user1
		for _, post := range user1Posts {
			assert.Equal(t, user1.ID, post.AuthorID)
		}

		// Test filtering by user2
		options.AuthorID = &user2.ID
		user2Posts, err := repo.List(options)
		assert.NoError(t, err)
		assert.Len(t, user2Posts, 2)

		// Verify all posts belong to user2
		for _, post := range user2Posts {
			assert.Equal(t, user2.ID, post.AuthorID)
		}

		// Test filtering by non-existent user
		nonExistentUserID := NonExistentUserID
		options.AuthorID = &nonExistentUserID
		emptyPosts, err := repo.List(options)
		assert.NoError(t, err)
		assert.Len(t, emptyPosts, 0)
	})
}

func TestCheckPermission(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		tx := setup()
		defer teardown(tx)

		// Create user and post
		user := firstCreateTestUser(t, tx, nil)
		repo := repository.NewPostRepositoryWithDB(tx)
		post := createTestPost(user.ID)
		createdPost, err := repo.Create(post)
		assert.NoError(t, err)

		// run
		err = repo.CheckPermission(createdPost.ID, user.ID)

		// assert
		assert.NoError(t, err)
	})

	t.Run("User does not own post", func(t *testing.T) {
		tx := setup()
		defer teardown(tx)

		// Create two users
		user1 := firstCreateTestUser(t, tx, nil)
		user2 := firstCreateTestUser(t, tx, map[string]interface{}{
			"username": "user2",
			"email":    "user2@test.com",
		})

		// Create post by user1
		repo := repository.NewPostRepositoryWithDB(tx)
		post := createTestPost(user1.ID)
		createdPost, err := repo.Create(post)
		assert.NoError(t, err)

		// run
		err = repo.CheckPermission(createdPost.ID, user2.ID)

		// assert
		assert.ErrorIs(t, err, apperrors.ErrForbidden)
	})
}
