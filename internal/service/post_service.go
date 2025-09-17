package service

import (
	"go-gin-api-server/internal/model"
	"go-gin-api-server/internal/repository"
	"go-gin-api-server/pkg/apperrors"
	"strconv"
	"strings"
)

type PostService interface {
	Create(post *model.Post) (*model.Post, error)
	List(request model.CursorRequest) (*model.CursorResponse[model.Post], error)
	GetByID(id uint64) (*model.Post, error)
	Update(id uint64, post *model.Post, currentUserID string) (*model.Post, error)
	Delete(id uint64, currentUserID string) error
}

type postServiceImpl struct {
	repo repository.PostRepository
}

func NewPostService(repo repository.PostRepository) PostService {
	return &postServiceImpl{repo: repo}
}

func (s *postServiceImpl) Create(post *model.Post) (*model.Post, error) {
	// business logic: validate content
	if err := s.validateContent(post.Content); err != nil {
		return nil, err
	}

	// business logic: validate sensitive words
	if s.containsSensitiveWords(post.Content) {
		return nil, apperrors.ErrPostContentSensitiveWords
	}

	return s.repo.Create(post)
}

func (s *postServiceImpl) List(request model.CursorRequest) (*model.CursorResponse[model.Post], error) {
	// Set defaults
	request.SetDefaults()

	// decode cursor
	var cursor model.Cursor
	if request.Cursor != "" {
		var err error
		cursor, err = model.DecodeCursor(request.Cursor)
		if err != nil {
			return nil, apperrors.ErrValidation
		}
	}

	opts := model.PostListOptions{
		Limit:    request.Limit + 1, // Request one extra to check if there are more results
		AuthorID: request.AuthorID,
		Cursor:   cursor,
	}

	posts, err := s.repo.List(opts)
	if err != nil {
		return nil, err
	}

	// Check if there are more results
	hasMore := len(posts) > request.Limit
	if hasMore {
		posts = posts[:request.Limit] // Remove the extra item
	}

	// Generate next cursor from the last item
	var nextCursor string
	if hasMore && len(posts) > 0 {
		lastPost := posts[len(posts)-1]
		nextCursor = model.EncodeCursor(model.Cursor{
			ID:        strconv.FormatUint(lastPost.ID, 10),
			CreatedAt: lastPost.CreatedAt,
		})
	}

	return &model.CursorResponse[model.Post]{
		Data:    posts,
		Next:    nextCursor,
		HasMore: hasMore,
	}, nil
}

func (s *postServiceImpl) GetByID(id uint64) (*model.Post, error) {
	return s.repo.FindByID(id)
}

func (s *postServiceImpl) Update(id uint64, post *model.Post, currentUserID string) (*model.Post, error) {
	// business logic: validate permission
	if err := s.repo.CheckPermission(id, currentUserID); err != nil {
		return nil, err
	}

	// business logic: validate content
	if post.Content != "" {
		if err := s.validateContent(post.Content); err != nil {
			return nil, err
		}

		if s.containsSensitiveWords(post.Content) {
			return nil, apperrors.ErrPostContentSensitiveWords
		}
	}

	return s.repo.Update(id, post)
}

func (s *postServiceImpl) Delete(id uint64, currentUserID string) error {
	// business logic: validate permission
	if err := s.repo.CheckPermission(id, currentUserID); err != nil {
		return err
	}

	return s.repo.Delete(id)
}

// business logic validation helper methods

func (s *postServiceImpl) validateContent(content string) error {
	content = strings.TrimSpace(content)

	if len(content) < 10 {
		return apperrors.ErrPostContentTooShort
	}

	if len(content) > 255 {
		return apperrors.ErrPostContentTooLong
	}

	return nil
}

func (s *postServiceImpl) containsSensitiveWords(content string) bool {
	sensitiveWords := []string{
		"violence",
	}

	content = strings.ToLower(content)
	for _, word := range sensitiveWords {
		if strings.Contains(content, word) {
			return true
		}
	}

	return false
}
