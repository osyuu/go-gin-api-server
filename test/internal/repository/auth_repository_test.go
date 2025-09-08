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
		repo := repository.NewAuthRepository()
		credentials := createTestCredentials()

		// run
		created, _ := repo.CreateCredentials(credentials)
		found, err := repo.FindByUserID(created.UserID)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, created, found)
	})

	t.Run("NotFound", func(t *testing.T) {
		repo := repository.NewAuthRepository()

		found, err := repo.FindByUserID("non-existent-user")

		assert.ErrorIs(t, err, apperrors.ErrNotFound)
		assert.Nil(t, found)
	})
}

func TestCreateCredentials(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo := repository.NewAuthRepository()
		credentials := createTestCredentials()

		// run
		created, err := repo.CreateCredentials(credentials)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, credentials, created)
	})

	t.Run("ErrUserExists", func(t *testing.T) {
		repo := repository.NewAuthRepository()
		credentials := createTestCredentials()

		// run
		_, _ = repo.CreateCredentials(credentials)
		existing, err := repo.CreateCredentials(credentials)

		// assert
		assert.ErrorIs(t, err, apperrors.ErrUserExists)
		assert.Nil(t, existing)
	})
}

func TestUpdatePassword(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo := repository.NewAuthRepository()
		credentials := createTestCredentials()

		// run
		created, _ := repo.CreateCredentials(credentials)
		err := repo.UpdatePassword(created.UserID, "new_password")
		found, _ := repo.FindByUserID(created.UserID)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, "new_password", found.Password)
	})

	t.Run("NotFound", func(t *testing.T) {
		repo := repository.NewAuthRepository()

		err := repo.UpdatePassword("non-existent-user", "new_password")

		assert.ErrorIs(t, err, apperrors.ErrNotFound)
	})
}

func TestDeleteCredentials(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo := repository.NewAuthRepository()
		credentials := createTestCredentials()

		// run
		created, _ := repo.CreateCredentials(credentials)
		err := repo.DeleteCredentials(created.UserID)
		found, err2 := repo.FindByUserID(created.UserID)

		// assert
		assert.NoError(t, err)
		assert.ErrorIs(t, err2, apperrors.ErrNotFound)
		assert.Nil(t, found)
	})

	t.Run("NotFound", func(t *testing.T) {
		repo := repository.NewAuthRepository()

		err := repo.DeleteCredentials("non-existent-user")

		assert.ErrorIs(t, err, apperrors.ErrNotFound)
	})
}

func TestAuthConcurrentAccess(t *testing.T) {
	repo := repository.NewAuthRepository()

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
		repo := repository.NewAuthRepository()
		credentials := &model.UserCredentials{
			UserID:   "",
			Password: "password",
		}
		_, err := repo.CreateCredentials(credentials)
		assert.ErrorIs(t, err, apperrors.ErrValidation) // Should return validation error
	})
}
