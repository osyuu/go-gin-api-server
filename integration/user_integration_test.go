package integration

import (
	"encoding/json"
	"go-gin-api-server/internal/handler"
	"go-gin-api-server/internal/repository"
	"go-gin-api-server/internal/service"
	"go-gin-api-server/pkg/utils"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupIntegrationUserRouter(db *gorm.DB) *gin.Engine {
	repo := repository.NewUserRepositoryWithDB(db)
	userService := service.NewUserService(repo)
	userHandler := handler.NewUserHandler(userService)
	r := gin.Default()

	// Register custom validator
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("username", utils.UsernameValidator)
	}

	r.GET("/users/:id", userHandler.GetUserByID)
	r.POST("/users", userHandler.CreateUser)
	return r
}

func TestIntegration_UserService_Success(t *testing.T) {
	db := setup()
	defer teardown(db)

	gin.SetMode(gin.TestMode)

	r := setupIntegrationUserRouter(db)
	ts := httptest.NewServer(r)
	defer ts.Close()

	// 1. create a user
	createResp, err := http.Post(ts.URL+"/users", "application/json",
		strings.NewReader(`{"name":"John Doe","username":"john_doe","email":"john@example.com"}`))

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, createResp.StatusCode)

	// 2. parse the created user response to get the actual user ID
	var createdUser map[string]interface{}
	err = json.NewDecoder(createResp.Body).Decode(&createdUser)
	createResp.Body.Close()
	assert.NoError(t, err)

	userID, ok := createdUser["id"].(string)
	assert.True(t, ok, "User ID should be a string")
	assert.NotEmpty(t, userID, "User ID should not be empty")

	// 3. query the user using the actual user ID
	resp, err := http.Get(ts.URL + "/users/" + userID)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// 4. check the response data
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	// Verify the response contains the expected user data
	var responseUser map[string]interface{}
	err = json.Unmarshal(body, &responseUser)
	assert.NoError(t, err)

	assert.Equal(t, userID, responseUser["id"])
	assert.Equal(t, "John Doe", responseUser["name"])
	assert.Equal(t, "john_doe", responseUser["username"])
	assert.Equal(t, "john@example.com", responseUser["email"])
	assert.Equal(t, true, responseUser["is_active"])
}

func TestIntegration_UserService_NotFound(t *testing.T) {
	db := setup()
	defer teardown(db)

	gin.SetMode(gin.TestMode)

	r := setupIntegrationUserRouter(db)
	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/users/999")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	// check response body
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	assert.JSONEq(t, `{"error":"User not found"}`, string(body))
}
