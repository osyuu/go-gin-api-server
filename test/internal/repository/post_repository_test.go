package repository

import (
	"blog_server/internal/model"
	"blog_server/internal/repository"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createTestPost() *model.Post {
	return &model.Post{
		ID: "1", Title: "Test Post", Content: "Test Content", Author: "Test Author",
	}
}

func createAllPosts() []model.Post {
	return []model.Post{
		{ID: "1", Title: "Test Post", Content: "Test Content", Author: "Test Author"},
		{ID: "2", Title: "Test Post 2", Content: "Test Content 2", Author: "Test Author 2"},
	}
}

// TestCases

func TestCreate_EmptyID(t *testing.T) {
	repo := repository.NewPostRepository()
	post := &model.Post{
		Title:   "Test Post",
		Content: "Test Content",
		Author:  "Test Author",
	}

	created, err := repo.Create(post)

	assert.NoError(t, err)
	assert.NotEmpty(t, created.ID)
}

func TestCreateAndFindByID(t *testing.T) {
	repo := repository.NewPostRepository()
	post := createTestPost()

	// run
	repo.Create(post)
	found, err := repo.FindByID("1")

	// assert
	assert.NoError(t, err)
	assert.Equal(t, post, found)
}

func TestCreateAndFindAll(t *testing.T) {
	repo := repository.NewPostRepository()
	posts := createAllPosts()

	// run
	repo.Create(&posts[0])
	repo.Create(&posts[1])
	found, err := repo.FindAll()

	// assert
	assert.NoError(t, err)
	assert.Equal(t, posts, found)
}

func TestUpdate_Success(t *testing.T) {
	repo := repository.NewPostRepository()
	post := createTestPost()

	// run
	repo.Create(post)
	update := &model.Post{Title: "Updated Title"}
	found, err := repo.Update("1", update)

	// assert
	assert.NoError(t, err)
	assert.Equal(t, "Updated Title", found.Title)
}

func TestUpdate_NotFound(t *testing.T) {
	repo := repository.NewPostRepository()
	post := createTestPost()

	// run
	repo.Create(post)
	update := &model.Post{Title: "Updated Title"}
	found, err := repo.Update("999", update)

	// assert
	assert.ErrorIs(t, err, repository.ErrNotFound)
	assert.Nil(t, found)
}

func TestDelete_Success(t *testing.T) {
	repo := repository.NewPostRepository()
	post := createTestPost()

	// run
	repo.Create(post)
	repo.Delete("1")
	found, err := repo.FindByID("1")

	// assert
	assert.ErrorIs(t, err, repository.ErrNotFound)
	assert.Nil(t, found)
}

func TestDelete_NotFound(t *testing.T) {
	repo := repository.NewPostRepository()

	// run
	err := repo.Delete("999")

	// assert
	assert.ErrorIs(t, err, repository.ErrNotFound)
}

func TestConcurrentAccess(t *testing.T) {
	repo := repository.NewPostRepository()

	// concurrent create and read
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			post := &model.Post{
				ID:    fmt.Sprintf("%d", id),
				Title: fmt.Sprintf("Post %d", id),
			}
			repo.Create(post)
		}(i)
	}
	wg.Wait()

	// verify all data are created correctly
	posts, err := repo.FindAll()
	assert.NoError(t, err)
	assert.Len(t, posts, 10)
}
