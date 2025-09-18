package handler

import (
	"go-gin-api-server/internal/model"
	"go-gin-api-server/pkg/apperrors"

	"github.com/gin-gonic/gin"
)

func GetUserID(c *gin.Context) (string, error) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		return "", apperrors.ErrUnauthorized
	}

	currentUserID, ok := userIDValue.(string)
	if !ok {
		return "", apperrors.ErrUnauthorized
	}

	return currentUserID, nil
}

func GetUserRole(c *gin.Context) (model.UserRole, error) {
	role, exists := c.Get("user_role")
	if !exists {
		return "", apperrors.ErrUnauthorized
	}

	userRole, ok := role.(model.UserRole)
	if !ok {
		return "", apperrors.ErrUnauthorized
	}

	return userRole, nil
}

func GetUserIDAndRole(c *gin.Context) (string, model.UserRole, error) {
	userID, err := GetUserID(c)
	if err != nil {
		return "", "", err
	}

	role, err := GetUserRole(c)
	if err != nil {
		return "", "", err
	}

	return userID, role, nil
}
