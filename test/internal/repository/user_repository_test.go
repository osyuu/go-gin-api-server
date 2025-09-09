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
		tx := setup()
		defer teardown(tx)

		repo := repository.NewUserRepositoryWithDB(tx)
		user := createTestUser()

		// run
		created, _ := repo.Create(user)
		found, err := repo.FindByID(created.ID)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, created.ID, found.ID)
		assert.Equal(t, created.Name, found.Name)
		assert.Equal(t, created.Username, found.Username)
		assert.Equal(t, created.Email, found.Email)
		assert.Equal(t, created.IsActive, found.IsActive)
		if created.BirthDate != nil && found.BirthDate != nil {
			assert.Equal(t, created.BirthDate.UTC(), found.BirthDate.UTC())
		} else {
			assert.Equal(t, created.BirthDate, found.BirthDate)
		}
		assert.Equal(t, created.CreatedAt.UTC(), found.CreatedAt.UTC())
		assert.Equal(t, created.UpdatedAt.UTC(), found.UpdatedAt.UTC())
	})

	t.Run("NotFound", func(t *testing.T) {
		tx := setup()
		defer teardown(tx)

		repo := repository.NewUserRepositoryWithDB(tx)

		found, err := repo.FindByID("non-existent-id")

		assert.ErrorIs(t, err, apperrors.ErrNotFound)
		assert.Nil(t, found)
	})
}

func TestCreateUserAndFindByUsername(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		tx := setup()
		defer teardown(tx)

		repo := repository.NewUserRepositoryWithDB(tx)
		user := createTestUser()

		// run
		created, _ := repo.Create(user)
		found, err := repo.FindByUsername(created.Username)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, created.ID, found.ID)
		assert.Equal(t, created.Name, found.Name)
		assert.Equal(t, created.Username, found.Username)
		assert.Equal(t, created.Email, found.Email)
		if created.BirthDate != nil && found.BirthDate != nil {
			assert.Equal(t, created.BirthDate.UTC(), found.BirthDate.UTC())
		} else {
			assert.Equal(t, created.BirthDate, found.BirthDate)
		}
		assert.Equal(t, created.CreatedAt.UTC(), found.CreatedAt.UTC())
		assert.Equal(t, created.UpdatedAt.UTC(), found.UpdatedAt.UTC())
	})

	t.Run("NotFound", func(t *testing.T) {
		tx := setup()
		defer teardown(tx)

		repo := repository.NewUserRepositoryWithDB(tx)

		found, err := repo.FindByUsername("non-existent-username")

		assert.ErrorIs(t, err, apperrors.ErrNotFound)
		assert.Nil(t, found)
	})
}

func TestCreateUserAndFindByEmail(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		tx := setup()
		defer teardown(tx)

		repo := repository.NewUserRepositoryWithDB(tx)
		user := createTestUser()

		// run
		created, _ := repo.Create(user)
		found, err := repo.FindByEmail(created.Email)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, created.ID, found.ID)
		assert.Equal(t, created.Name, found.Name)
		assert.Equal(t, created.Username, found.Username)
		assert.Equal(t, created.Email, found.Email)
		if created.BirthDate != nil && found.BirthDate != nil {
			assert.Equal(t, created.BirthDate.UTC(), found.BirthDate.UTC())
		} else {
			assert.Equal(t, created.BirthDate, found.BirthDate)
		}
		assert.Equal(t, created.CreatedAt.UTC(), found.CreatedAt.UTC())
		assert.Equal(t, created.UpdatedAt.UTC(), found.UpdatedAt.UTC())
	})

	t.Run("NotFound", func(t *testing.T) {
		tx := setup()
		defer teardown(tx)

		repo := repository.NewUserRepositoryWithDB(tx)

		found, err := repo.FindByEmail("non-existent-email")

		assert.ErrorIs(t, err, apperrors.ErrNotFound)
		assert.Nil(t, found)
	})
}

func TestUpdateUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		tx := setup()
		defer teardown(tx)

		repo := repository.NewUserRepositoryWithDB(tx)
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
		assert.Equal(t, updated.Username, found.Username)
		assert.Equal(t, updated.Email, found.Email)
		if updated.BirthDate != nil && found.BirthDate != nil {
			assert.Equal(t, updated.BirthDate.UTC(), found.BirthDate.UTC())
		} else {
			assert.Equal(t, updated.BirthDate, found.BirthDate)
		}
		assert.Equal(t, created.CreatedAt.UTC(), found.CreatedAt.UTC())
		assert.True(t, found.UpdatedAt.After(found.CreatedAt))
	})

	t.Run("NotFound", func(t *testing.T) {
		tx := setup()
		defer teardown(tx)

		repo := repository.NewUserRepositoryWithDB(tx)

		updated := &model.User{Name: "updated"}
		found, err := repo.Update("non-existent-id", updated)

		assert.ErrorIs(t, err, apperrors.ErrNotFound)
		assert.Nil(t, found)
	})
}

func TestCreateUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		tx := setup()
		defer teardown(tx)

		repo := repository.NewUserRepositoryWithDB(tx)
		user := createTestUser()

		// run
		created, err := repo.Create(user)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, user.ID, created.ID)
		assert.Equal(t, user.Name, created.Name)
		assert.Equal(t, user.Username, created.Username)
		assert.Equal(t, user.Email, created.Email)
		if user.BirthDate != nil && created.BirthDate != nil {
			assert.Equal(t, user.BirthDate.UTC(), created.BirthDate.UTC())
		} else {
			assert.Equal(t, user.BirthDate, created.BirthDate)
		}
		assert.Equal(t, user.CreatedAt.UTC(), created.CreatedAt.UTC())
		assert.Equal(t, user.UpdatedAt.UTC(), created.UpdatedAt.UTC())
	})

	t.Run("ErrUserExists", func(t *testing.T) {
		tx := setup()
		defer teardown(tx)

		repo := repository.NewUserRepositoryWithDB(tx)
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
		tx := setup()
		defer teardown(tx)

		repo := repository.NewUserRepositoryWithDB(tx)
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
		tx := setup()
		defer teardown(tx)

		repo := repository.NewUserRepositoryWithDB(tx)

		err := repo.Delete("non-existent-id")

		assert.ErrorIs(t, err, apperrors.ErrNotFound)
	})

}
