package repository

import (
	"fmt"
	"go-gin-api-server/internal/model"
	"go-gin-api-server/internal/repository"
	"go-gin-api-server/pkg/apperrors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createTestCredentials() *model.UserCredentials {
	return &model.UserCredentials{
		UserID:   "user123",
		Password: "hashed_password",
	}
}

// Testcases

func TestCreateCredentialsAndFindByUserID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		tx := setup()
		defer teardown(tx)
		repo := repository.NewAuthRepositoryWithDB(tx)
		credentials := createTestCredentials()

		// run
		created, err := repo.CreateCredentials(credentials)
		assert.NoError(t, err)

		found, err := repo.FindByUserID(created.UserID)
		assert.NoError(t, err)

		// assert
		assert.Equal(t, created.UserID, found.UserID)
		assert.Equal(t, created.Password, found.Password)
		assert.Equal(t, created.CreatedAt.UTC(), found.CreatedAt.UTC())
		assert.Equal(t, created.UpdatedAt.UTC(), found.UpdatedAt.UTC())
	})

	t.Run("NotFound", func(t *testing.T) {
		tx := setup()
		defer teardown(tx)

		repo := repository.NewAuthRepositoryWithDB(tx)

		found, err := repo.FindByUserID("non-existent-user")

		assert.ErrorIs(t, err, apperrors.ErrNotFound)
		assert.Nil(t, found)
	})
}

func TestCreateCredentials(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		tx := setup()
		defer teardown(tx)
		repo := repository.NewAuthRepositoryWithDB(tx)
		credentials := createTestCredentials()

		// run
		created, err := repo.CreateCredentials(credentials)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, credentials.UserID, created.UserID)
		assert.Equal(t, credentials.Password, created.Password)
		assert.NotZero(t, created.CreatedAt)
		assert.NotZero(t, created.UpdatedAt)
	})

	t.Run("ErrUserExists", func(t *testing.T) {
		tx := setup()
		defer teardown(tx)
		repo := repository.NewAuthRepositoryWithDB(tx)
		credentials := createTestCredentials()

		// run
		_, err := repo.CreateCredentials(credentials)
		assert.NoError(t, err)
		existing, err := repo.CreateCredentials(credentials)

		// assert
		assert.ErrorIs(t, err, apperrors.ErrUserExists)
		assert.Nil(t, existing)
	})
}

func TestUpdatePassword(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		tx := setup()
		defer teardown(tx)
		repo := repository.NewAuthRepositoryWithDB(tx)
		credentials := createTestCredentials()

		// run
		created, err := repo.CreateCredentials(credentials)
		assert.NoError(t, err)
		err = repo.UpdatePassword(created.UserID, "new_password")
		assert.NoError(t, err)
		found, err := repo.FindByUserID(created.UserID)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, created.UserID, found.UserID)
		assert.Equal(t, created.CreatedAt.UTC(), found.CreatedAt.UTC())
		assert.Equal(t, "new_password", found.Password)
		assert.True(t, found.UpdatedAt.After(found.CreatedAt))
	})

	t.Run("NotFound", func(t *testing.T) {
		tx := setup()
		defer teardown(tx)
		repo := repository.NewAuthRepositoryWithDB(tx)

		err := repo.UpdatePassword("non-existent-user", "new_password")

		assert.ErrorIs(t, err, apperrors.ErrNotFound)
	})
}

func TestDeleteCredentials(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		tx := setup()
		defer teardown(tx)
		repo := repository.NewAuthRepositoryWithDB(tx)
		credentials := createTestCredentials()

		// run
		created, err := repo.CreateCredentials(credentials)
		assert.NoError(t, err)
		err = repo.DeleteCredentials(created.UserID)
		found, err2 := repo.FindByUserID(created.UserID)

		// assert
		assert.NoError(t, err)
		assert.ErrorIs(t, err2, apperrors.ErrNotFound)
		assert.Nil(t, found)
	})

	t.Run("NotFound", func(t *testing.T) {
		tx := setup()
		defer teardown(tx)
		repo := repository.NewAuthRepositoryWithDB(tx)

		err := repo.DeleteCredentials("non-existent-user")

		assert.ErrorIs(t, err, apperrors.ErrNotFound)
	})
}

func TestAuthConcurrentAccess(t *testing.T) {
	tx := setup()
	defer teardown(tx)
	repo := repository.NewAuthRepositoryWithDB(tx)

	// Test concurrent creation
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			credentials := &model.UserCredentials{
				UserID:   fmt.Sprintf("user%d", id),
				Password: "password",
			}
			_, err := repo.CreateCredentials(credentials)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for range 10 {
		<-done
	}

	// Verify all credentials were created
	for i := 0; i < 10; i++ {
		_, err := repo.FindByUserID(fmt.Sprintf("user%d", i))
		assert.NoError(t, err)
	}
}

func TestAuthEdgeCases(t *testing.T) {
	t.Run("EmptyUserID", func(t *testing.T) {
		tx := setup()
		defer teardown(tx)
		repo := repository.NewAuthRepositoryWithDB(tx)
		credentials := &model.UserCredentials{
			UserID:   "",
			Password: "password",
		}
		_, err := repo.CreateCredentials(credentials)
		assert.ErrorIs(t, err, apperrors.ErrValidation) // Should return validation error
	})
}
