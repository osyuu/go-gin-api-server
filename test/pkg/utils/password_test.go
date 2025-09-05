package utils

import (
	"go-gin-api-server/pkg/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashPassword(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		password := "testpassword123"

		hashedPassword, err := utils.HashPassword(password)

		assert.NoError(t, err)
		assert.NotEmpty(t, hashedPassword)
		assert.NotEqual(t, password, hashedPassword)
	})
}

func TestCheckPassword(t *testing.T) {
	t.Run("ValidPassword", func(t *testing.T) {
		password := "testpassword123"
		hashedPassword, err := utils.HashPassword(password)
		assert.NoError(t, err)

		err = utils.CheckPassword(hashedPassword, password)

		assert.NoError(t, err)
	})

	t.Run("InvalidPassword", func(t *testing.T) {
		password := "testpassword123"
		wrongPassword := "wrongpassword"
		hashedPassword, err := utils.HashPassword(password)
		assert.NoError(t, err)

		err = utils.CheckPassword(hashedPassword, wrongPassword)

		assert.Error(t, err)
	})
}

func TestPasswordRoundTrip(t *testing.T) {
	t.Run("HashAndCheck", func(t *testing.T) {
		password := "testpassword123"

		hashedPassword, err := utils.HashPassword(password)
		assert.NoError(t, err)

		err = utils.CheckPassword(hashedPassword, password)
		assert.NoError(t, err)
	})
}
