package repository

import (
	"go-gin-api-server/internal/model"
	"go-gin-api-server/internal/repository"
	"go-gin-api-server/pkg/apperrors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func createTestUser() *model.User {
	birthDate := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	return model.CreateUser(
		"test",
		"test",
		"test@test.com",
		&birthDate,
	)
}

// Testcases

func TestCreateUserAndFindByID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo := repository.NewUserRepository()
		user := createTestUser()

		// run
		created, _ := repo.Create(user)
		found, err := repo.FindByID(created.ID)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, created, found)
	})

	t.Run("NotFound", func(t *testing.T) {
		repo := repository.NewUserRepository()

		found, err := repo.FindByID("non-existent-id")

		assert.ErrorIs(t, err, apperrors.ErrNotFound)
		assert.Nil(t, found)
	})
}

func TestCreateUserAndFindByUsername(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo := repository.NewUserRepository()
		user := createTestUser()

		// run
		created, _ := repo.Create(user)
		found, err := repo.FindByUsername(created.Username)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, created, found)
	})

	t.Run("NotFound", func(t *testing.T) {
		repo := repository.NewUserRepository()

		found, err := repo.FindByUsername("non-existent-username")

		assert.ErrorIs(t, err, apperrors.ErrNotFound)
		assert.Nil(t, found)
	})
}

func TestCreateUserAndFindByEmail(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo := repository.NewUserRepository()
		user := createTestUser()

		// run
		created, _ := repo.Create(user)
		found, err := repo.FindByEmail(created.Email)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, created, found)
	})

	t.Run("NotFound", func(t *testing.T) {
		repo := repository.NewUserRepository()

		found, err := repo.FindByEmail("non-existent-email")

		assert.ErrorIs(t, err, apperrors.ErrNotFound)
		assert.Nil(t, found)
	})
}

func TestUpdateUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo := repository.NewUserRepository()
		user := createTestUser()

		// run
		created, _ := repo.Create(user)
		birthDate := time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)
		updated := &model.User{
			Name:      "updated",
			Username:  "updated",
			Email:     "updated@test.com",
			BirthDate: &birthDate,
		}
		found, err := repo.Update(created.ID, updated)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, created.ID, found.ID)
		assert.Equal(t, updated.Name, found.Name)
		assert.Equal(t, updated.BirthDate, found.BirthDate)
		assert.Equal(t, updated.Username, found.Username)
		assert.Equal(t, updated.Email, found.Email)
		assert.Equal(t, created.CreatedAt, found.CreatedAt)
		assert.True(t, found.UpdatedAt.After(found.CreatedAt))
	})

	t.Run("NotFound", func(t *testing.T) {
		repo := repository.NewUserRepository()

		updated := &model.User{Name: "updated"}
		found, err := repo.Update("non-existent-id", updated)

		assert.ErrorIs(t, err, apperrors.ErrNotFound)
		assert.Nil(t, found)
	})
}

func TestCreateUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo := repository.NewUserRepository()
		user := createTestUser()

		// run
		created, err := repo.Create(user)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, user, created)
	})

	t.Run("ErrUserExists", func(t *testing.T) {
		repo := repository.NewUserRepository()
		user := createTestUser()
		user2 := createTestUser()

		// run
		_, _ = repo.Create(user)
		existing, err := repo.Create(user2)

		// assert
		assert.ErrorIs(t, err, apperrors.ErrUserExists)
		assert.Nil(t, existing)
	})
}

func TestDeleteUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo := repository.NewUserRepository()
		user := createTestUser()

		// run
		created, _ := repo.Create(user)
		repo.Delete(created.ID)
		found, err := repo.FindByID(created.ID)

		// assert
		assert.ErrorIs(t, err, apperrors.ErrNotFound)
		assert.Nil(t, found)
	})

	t.Run("NotFound", func(t *testing.T) {
		repo := repository.NewUserRepository()

		err := repo.Delete("non-existent-id")

		assert.ErrorIs(t, err, apperrors.ErrNotFound)
	})

}
