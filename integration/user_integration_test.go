package integration

import (
	"blog_server/internal/handler"
	"blog_server/internal/repository"
	"blog_server/internal/service"
	"blog_server/pkg/utils"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func setupIntegrationUserRouter() *gin.Engine {
	repo := repository.NewUserRepository()
	userService := service.NewUserService(repo)
	userHandler := handler.NewUserHandler(userService)
	r := gin.Default()

	// Register custom validator
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("custom_validator", utils.CustomValidator)
	}

	r.GET("/users/:id", userHandler.GetUser)
	r.POST("/users", userHandler.CreateUser)
	return r
}

func TestIntegration_UserService_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := setupIntegrationUserRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()

	// 1. create a user
	createResp, err := http.Post(ts.URL+"/users", "application/json",
		strings.NewReader(`{"name":"John Doe","age":20}`))

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, createResp.StatusCode)
	createResp.Body.Close()

	// 2. query the user
	resp, err := http.Get(ts.URL + "/users/1")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// 3. check the response data
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	assert.JSONEq(t, `{"id":"1","name":"John Doe","age":20}`, string(body))
}

func TestIntegration_UserService_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := setupIntegrationUserRouter()
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
