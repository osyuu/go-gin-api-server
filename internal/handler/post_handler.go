package handler

import (
	"errors"
	"go-gin-api-server/internal/middleware"
	"go-gin-api-server/internal/model"
	"go-gin-api-server/internal/repository"
	"go-gin-api-server/internal/service"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PostHandler struct {
	service service.PostService
}

func NewPostHandler(service service.PostService) *PostHandler {
	return &PostHandler{service: service}
}

func (h *PostHandler) RegisterRoutes(r *gin.Engine) {
	router := r.Group("/api/v1")
	{
		router.GET("/posts", h.GetPosts)
		router.GET("/posts/:id", h.GetPostByID)
	}
}

func (h *PostHandler) RegisterProtectedRoutes(r *gin.Engine, authMiddleware *middleware.AuthMiddleware) {
	protected := r.Group("/api/v1/posts")
	protected.Use(authMiddleware.RequireAuth())
	{
		protected.POST("", h.CreatePost)
		protected.PATCH("/:id", h.UpdatePost)
		protected.DELETE("/:id", h.DeletePost)
	}
}

func (h *PostHandler) GetPosts(c *gin.Context) {
	posts, err := h.service.GetAll()
	if err != nil {
		log.Printf("Error getting posts: %v", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve posts",
		})
		return
	}

	c.JSON(http.StatusOK, posts)
}

func (h *PostHandler) GetPostByID(c *gin.Context) {
	id := c.Param("id")

	found, err := h.service.GetByID(id)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			log.Printf("Post not found: %s", id)
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Post not found",
			})
		default:
			log.Printf("Unexpected error getting post %s: %v", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get post",
			})
		}
		return
	}

	c.JSON(http.StatusOK, found)
}

func (h *PostHandler) CreatePost(c *gin.Context) {
	var newPost model.Post

	if err := c.ShouldBindJSON(&newPost); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
		})
		return
	}

	created, err := h.service.Create(&newPost)
	if err != nil {
		log.Printf("Error creating post: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create post",
		})
		return
	}

	c.JSON(http.StatusCreated, created)
}

func (h *PostHandler) UpdatePost(c *gin.Context) {
	id := c.Param("id")
	var update model.Post

	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
		})
		return
	}

	updated, err := h.service.Update(id, &update)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			log.Printf("Post not found: %s", id)
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Post not found",
			})
		default:
			log.Printf("Unexpected error updating post %s: %v", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update post",
			})
		}
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h *PostHandler) DeletePost(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.Delete(id); err != nil {

		switch {
		case errors.Is(err, repository.ErrNotFound):
			log.Printf("Post not found: %s", id)
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Post not found",
			})
		default:
			log.Printf("Unexpected error deleting post %s: %v", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to delete post",
			})
		}
		return
	}

	c.Status(http.StatusNoContent)
}
