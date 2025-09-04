package service

import (
	"go-gin-api-server/internal/model"
	"go-gin-api-server/internal/repository"
)

type PostService interface {
	Create(post *model.Post) (*model.Post, error)
	GetAll() ([]model.Post, error)
	GetByID(id string) (*model.Post, error)
	Update(id string, post *model.Post) (*model.Post, error)
	Delete(id string) error
}

type postServiceImpl struct {
	repo repository.PostRepository
}

func NewPostService(repo repository.PostRepository) PostService {
	return &postServiceImpl{repo: repo}
}

func (s *postServiceImpl) Create(post *model.Post) (*model.Post, error) {
	return s.repo.Create(post)
}

func (s *postServiceImpl) GetAll() ([]model.Post, error) {
	return s.repo.FindAll()
}

func (s *postServiceImpl) GetByID(id string) (*model.Post, error) {
	return s.repo.FindByID(id)
}

func (s *postServiceImpl) Update(id string, post *model.Post) (*model.Post, error) {
	return s.repo.Update(id, post)
}

func (s *postServiceImpl) Delete(id string) error {
	return s.repo.Delete(id)
}
