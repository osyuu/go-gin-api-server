package repository

import (
	"errors"
	"go-gin-api-server/internal/database"
	"go-gin-api-server/internal/model"
	"go-gin-api-server/pkg/apperrors"
	"strconv"

	"gorm.io/gorm"
)

type PostRepository interface {
	Create(post *model.Post) (*model.Post, error)
	List(opts model.PostListOptions) ([]model.Post, error)
	FindByID(id uint64) (*model.Post, error)
	Update(id uint64, post *model.Post) (*model.Post, error)
	Delete(id uint64) error
	CheckPermission(id uint64, currentUserID string) error
}

type postRepositoryImpl struct {
	db *gorm.DB
}

func NewPostRepository() PostRepository {
	return &postRepositoryImpl{
		db: database.GetDB(),
	}
}

func NewPostRepositoryWithDB(db *gorm.DB) PostRepository {
	return &postRepositoryImpl{
		db: db,
	}
}

func (r *postRepositoryImpl) Create(post *model.Post) (*model.Post, error) {
	if err := r.db.Create(post).Error; err != nil {
		return nil, err
	}
	return post, nil
}

func (r *postRepositoryImpl) List(opts model.PostListOptions) ([]model.Post, error) {
	var posts []model.Post

	// validate negative
	if opts.Limit < 0 {
		return nil, apperrors.ErrValidation
	}

	query := r.db.Preload("Author", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, name, username")
	}).
		Order("created_at DESC, id DESC").
		Limit(opts.Limit)

	// handle cursor pagination
	if opts.Cursor.ID != "" {
		cursorID, err := strconv.ParseInt(opts.Cursor.ID, 10, 64)
		if err != nil {
			return nil, apperrors.ErrValidation
		}

		// only add WHERE condition when cursorID > 0
		if cursorID > 0 {
			query = query.Where("(created_at < ?) OR (created_at = ? AND id < ?)", opts.Cursor.CreatedAt, opts.Cursor.CreatedAt, cursorID)
		}
	}

	// add optional filter
	if opts.AuthorID != nil {
		query = query.Where("author_id = ?", *opts.AuthorID)
	}

	if err := query.Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *postRepositoryImpl) FindByID(id uint64) (*model.Post, error) {
	var post model.Post
	if err := r.db.Preload("Author", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, name, username")
	}).
		Where("id = ?", id).
		First(&post).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &post, nil
}

func (r *postRepositoryImpl) Update(id uint64, updated *model.Post) (*model.Post, error) {
	result := r.db.Model(&model.Post{}).
		Where("id = ?", id).
		Updates(updated)

	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.ErrNotFound
	}

	var post model.Post
	if err := r.db.Where("id = ?", id).
		First(&post).Error; err != nil {
		return nil, apperrors.ErrNotFound
	}
	return &post, nil
}

func (r *postRepositoryImpl) Delete(id uint64) error {
	result := r.db.Where("id = ?", id).Delete(&model.Post{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *postRepositoryImpl) CheckPermission(id uint64, userID string) error {
	var count int64
	err := r.db.Model(&model.Post{}).
		Where("id = ? AND author_id = ?", id, userID).
		Count(&count).Error

	if err != nil {
		return err
	}

	if count == 0 {
		return apperrors.ErrForbidden
	}

	return nil
}
