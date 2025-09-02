package repository

import (
	"blog_server/internal/model"
	"errors"
	"fmt"
	"sync"
)

var (
	ErrNotFound           = errors.New("post not found")
	ErrValidation         = errors.New("validation error")
	ErrDatabaseConnection = errors.New("database connection error")
)

type PostRepository interface {
	Create(post *model.Post) (*model.Post, error)
	FindAll() ([]model.Post, error)
	FindByID(id string) (*model.Post, error)
	Update(id string, post *model.Post) (*model.Post, error)
	Delete(id string) error
}

type postRepositoryImpl struct {
	mutex sync.RWMutex
	posts []model.Post
}

func NewPostRepository() PostRepository {
	return &postRepositoryImpl{
		posts: make([]model.Post, 0),
	}
}

func (r *postRepositoryImpl) Create(post *model.Post) (*model.Post, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if post.ID == "" {
		post.ID = fmt.Sprintf("%d", len(r.posts)+1)
	}

	r.posts = append(r.posts, *post)
	return post, nil
}

func (r *postRepositoryImpl) FindAll() ([]model.Post, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return append([]model.Post{}, r.posts...), nil
}

func (r *postRepositoryImpl) FindByID(id string) (*model.Post, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	for _, post := range r.posts {
		if post.ID == id {
			return &post, nil
		}
	}
	return nil, ErrNotFound
}

func (r *postRepositoryImpl) Update(id string, updated *model.Post) (*model.Post, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for i, post := range r.posts {
		if post.ID == id {
			if updated.Title != "" {
				r.posts[i].Title = updated.Title
			}
			if updated.Content != "" {
				r.posts[i].Content = updated.Content
			}
			if updated.Author != "" {
				r.posts[i].Author = updated.Author
			}
			return &r.posts[i], nil
		}
	}
	return nil, ErrNotFound
}

func (r *postRepositoryImpl) Delete(id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for i, post := range r.posts {
		if post.ID == id {
			r.posts = append(r.posts[:i], r.posts[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}
