package integration

import (
	"go-gin-api-server/internal/model"
	"go-gin-api-server/internal/repository"
	"go-gin-api-server/internal/service"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntegration_PostServiceLifecycle(t *testing.T) {
	repo := repository.NewPostRepository()
	service := service.NewPostService(repo)

	// 1. create a post
	post := &model.Post{
		ID:      "200",
		Title:   "Test Post",
		Content: "Test Content",
		Author:  "Test Author",
	}
	created, err := service.Create(post)
	assert.NoError(t, err)
	assert.Equal(t, post, created)

	// 2. get the post
	found, err := service.GetByID("200")
	assert.NoError(t, err)
	assert.Equal(t, created, found)

	// 3. update the post
	updated, err := service.Update("200", &model.Post{
		Title:   "Updated Title",
		Content: "Updated Content",
	})
	assert.NoError(t, err)
	assert.Equal(t, "Updated Title", updated.Title)
	assert.Equal(t, "Updated Content", updated.Content)

	// 4. find all posts
	posts, err := service.GetAll()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(posts))
	assert.Equal(t, "Updated Title", posts[0].Title)
	assert.Equal(t, "Updated Content", posts[0].Content)

	// 5. delete the post
	err = service.Delete("200")
	assert.NoError(t, err)

	// 6. confirm the post is deleted
	found, err = service.GetByID("200")
	assert.ErrorIs(t, err, repository.ErrNotFound)
	assert.Nil(t, found)
}
