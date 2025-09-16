package integration

import (
	"fmt"
	"go-gin-api-server/internal/handler"
	"go-gin-api-server/internal/middleware"
	"go-gin-api-server/internal/model"
	"go-gin-api-server/internal/repository"
	"go-gin-api-server/internal/service"
	"go-gin-api-server/pkg/logger"
	"go-gin-api-server/pkg/utils"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupIntegrationPostRouter(db *gorm.DB) *gin.Engine {
	// Setup dependencies (使用 TestMain 中初始化的全局變量)
	userRepo := repository.NewUserRepositoryWithDB(db)
	authRepo := repository.NewAuthRepositoryWithDB(db)
	postRepo := repository.NewPostRepositoryWithDB(db)

	// Setup services (使用全局的 JWT Manager)
	authService := service.NewAuthService(userRepo, authRepo, globalJWTManager)
	postService := service.NewPostService(postRepo)

	// Setup handlers and middleware
	postHandler := handler.NewPostHandler(postService, logger.Log)
	authMiddleware := middleware.NewAuthMiddleware(authService)

	// Setup router
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Register custom validator
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("username", utils.UsernameValidator)
	}

	// 註冊公開路由
	postHandler.RegisterRoutes(r)

	// 註冊受保護的路由
	postHandler.RegisterProtectedRoutes(r, authMiddleware)

	return r
}

// Helper function
func createTestUser(t *testing.T, db *gorm.DB, overrides ...map[string]interface{}) *model.User {
	username := "testuser"
	email := "test@example.com"

	if len(overrides) > 0 {
		override := overrides[0]
		if val, ok := override["username"]; ok {
			username = val.(string)
		}
		if val, ok := override["email"]; ok {
			email = val.(string)
		}
	}

	user := model.CreateUser(
		"Test User",
		&username,
		&email,
		nil,
	)

	userRepo := repository.NewUserRepositoryWithDB(db)
	createdUser, err := userRepo.Create(user)
	assert.NoError(t, err)
	return createdUser
}

func createTestToken(t *testing.T, user *model.User) *model.TokenResponse {
	token, err := globalJWTManager.GenerateToken(user)
	assert.NoError(t, err)
	return token
}

func TestPostIntegration_PostLifecycle(t *testing.T) {
	db := setup()
	defer teardown(db)
	router := setupIntegrationPostRouter(db)

	// 1. 創建用戶和認證
	user := createTestUser(t, db)
	token := createTestToken(t, user)
	accessToken := token.AccessToken

	// 2. 創建 post
	createReq := map[string]interface{}{
		"content": "Test Post Content",
	}
	createResp := makeHTTPRequest(t, router, "POST", "/api/v1/posts", createReq, accessToken)
	assert.Equal(t, 201, createResp.Code)

	var createdPost model.Post
	parseJSONResponse(t, createResp, &createdPost)
	assert.Equal(t, "Test Post Content", createdPost.Content)
	assert.Equal(t, user.ID, createdPost.AuthorID)
	postID := createdPost.ID

	// 3. 獲取 post
	getResp := makeHTTPRequest(t, router, "GET", fmt.Sprintf("/api/v1/posts/%d", postID), nil, "")
	assert.Equal(t, 200, getResp.Code)

	var retrievedPost model.Post
	parseJSONResponse(t, getResp, &retrievedPost)
	assert.Equal(t, createdPost.ID, retrievedPost.ID)
	assert.Equal(t, createdPost.Content, retrievedPost.Content)

	// 4. 更新 post
	updateReq := map[string]interface{}{
		"content": "Updated Post Content",
	}
	updateResp := makeHTTPRequest(t, router, "PATCH", fmt.Sprintf("/api/v1/posts/%d", postID), updateReq, accessToken)
	assert.Equal(t, 200, updateResp.Code)

	var updatedPost model.Post
	parseJSONResponse(t, updateResp, &updatedPost)
	assert.Equal(t, "Updated Post Content", updatedPost.Content)

	// 5. 刪除 post
	deleteResp := makeHTTPRequest(t, router, "DELETE", fmt.Sprintf("/api/v1/posts/%d", postID), nil, accessToken)
	assert.Equal(t, 204, deleteResp.Code)

	// 6. 確認 post 已刪除
	getDeletedResp := makeHTTPRequest(t, router, "GET", fmt.Sprintf("/api/v1/posts/%d", postID), nil, "")
	assert.Equal(t, 404, getDeletedResp.Code)
}

func TestPostIntegration_CreatePost_Unauthorized(t *testing.T) {
	db := setup()
	defer teardown(db)
	router := setupIntegrationPostRouter(db)

	// 測試未認證的創建請求
	createReq := map[string]interface{}{
		"content": "Test Post Content",
	}
	createResp := makeHTTPRequest(t, router, "POST", "/api/v1/posts", createReq, "")
	assert.Equal(t, 401, createResp.Code)
}

func TestPostIntegration_CreatePost_InvalidData(t *testing.T) {
	db := setup()
	defer teardown(db)
	router := setupIntegrationPostRouter(db)

	// 創建用戶和認證
	user := createTestUser(t, db)
	token := createTestToken(t, user)

	// 測試無效數據（空內容）
	createReq := map[string]interface{}{
		"content": "",
	}
	createResp := makeHTTPRequest(t, router, "POST", "/api/v1/posts", createReq, token.AccessToken)
	assert.Equal(t, 400, createResp.Code)
}

func TestPostIntegration_GetPost_NotFound(t *testing.T) {
	db := setup()
	defer teardown(db)
	router := setupIntegrationPostRouter(db)

	// 測試獲取不存在的 post
	getResp := makeHTTPRequest(t, router, "GET", fmt.Sprintf("/api/v1/posts/%d", NonExistentPostID), nil, "")
	assert.Equal(t, 404, getResp.Code)
}

func TestPostIntegration_UpdatePost_Unauthorized(t *testing.T) {
	db := setup()
	defer teardown(db)
	router := setupIntegrationPostRouter(db)

	// 創建用戶和 post
	user := createTestUser(t, db)
	token := createTestToken(t, user)

	createReq := map[string]interface{}{
		"content": "Test Post Content",
	}
	createResp := makeHTTPRequest(t, router, "POST", "/api/v1/posts", createReq, token.AccessToken)
	assert.Equal(t, 201, createResp.Code)

	var createdPost model.Post
	parseJSONResponse(t, createResp, &createdPost)

	// 測試未認證的更新請求
	updateReq := map[string]interface{}{
		"content": "Updated Content",
	}
	updateResp := makeHTTPRequest(t, router, "PATCH", fmt.Sprintf("/api/v1/posts/%d", createdPost.ID), updateReq, "")
	assert.Equal(t, 401, updateResp.Code)
}

func TestPostIntegration_GetPosts_CursorPagination(t *testing.T) {
	db := setup()
	defer teardown(db)
	router := setupIntegrationPostRouter(db)

	// 創建用戶和認證
	user := createTestUser(t, db)
	token := createTestToken(t, user)
	accessToken := token.AccessToken

	// 創建多個 posts 用於測試分頁
	postContents := []string{
		"Post 1 - First",
		"Post 2 - Second",
		"Post 3 - Third",
		"Post 4 - Fourth",
		"Post 5 - Fifth",
	}

	for _, content := range postContents {
		createReq := map[string]interface{}{
			"content": content,
		}
		createResp := makeHTTPRequest(t, router, "POST", "/api/v1/posts", createReq, accessToken)
		assert.Equal(t, 201, createResp.Code)
	}

	// 1. 測試第一頁（limit=2）
	getResp := makeHTTPRequest(t, router, "GET", "/api/v1/posts?limit=2", nil, "")
	assert.Equal(t, 200, getResp.Code)

	var firstPage model.CursorResponse[model.Post]
	parseJSONResponse(t, getResp, &firstPage)
	assert.Len(t, firstPage.Data, 2)
	assert.True(t, firstPage.HasMore)
	assert.NotEmpty(t, firstPage.Next)

	// 2. 測試第二頁（使用 cursor）
	getResp2 := makeHTTPRequest(t, router, "GET", fmt.Sprintf("/api/v1/posts?limit=2&cursor=%s", firstPage.Next), nil, "")
	assert.Equal(t, 200, getResp2.Code)

	var secondPage model.CursorResponse[model.Post]
	parseJSONResponse(t, getResp2, &secondPage)
	assert.Len(t, secondPage.Data, 2)
	assert.True(t, secondPage.HasMore)
	assert.NotEmpty(t, secondPage.Next)

	// 3. 測試最後一頁
	getResp3 := makeHTTPRequest(t, router, "GET", fmt.Sprintf("/api/v1/posts?limit=2&cursor=%s", secondPage.Next), nil, "")
	assert.Equal(t, 200, getResp3.Code)

	var lastPage model.CursorResponse[model.Post]
	parseJSONResponse(t, getResp3, &lastPage)
	assert.Len(t, lastPage.Data, 1) // 只剩一個 post
	assert.False(t, lastPage.HasMore)
	assert.Empty(t, lastPage.Next)

	// 4. 驗證數據順序（應該按 created_at DESC 排序）
	assert.Equal(t, "Post 5 - Fifth", firstPage.Data[0].Content)
	assert.Equal(t, "Post 4 - Fourth", firstPage.Data[1].Content)
	assert.Equal(t, "Post 3 - Third", secondPage.Data[0].Content)
	assert.Equal(t, "Post 2 - Second", secondPage.Data[1].Content)
	assert.Equal(t, "Post 1 - First", lastPage.Data[0].Content)
}

func TestPostIntegration_GetPosts_InvalidCursor(t *testing.T) {
	db := setup()
	defer teardown(db)
	router := setupIntegrationPostRouter(db)

	// 測試無效的 cursor
	getResp := makeHTTPRequest(t, router, "GET", "/api/v1/posts?limit=10&cursor=invalid_cursor", nil, "")
	assert.Equal(t, 400, getResp.Code)
}

func TestPostIntegration_GetPosts_AuthorFilter(t *testing.T) {
	db := setup()
	defer teardown(db)
	router := setupIntegrationPostRouter(db)

	// 創建兩個用戶
	user1 := createTestUser(t, db)
	user2 := createTestUser(t, db, map[string]interface{}{
		"username": "user2",
		"email":    "user2@example.com",
	})
	token1 := createTestToken(t, user1)
	token2 := createTestToken(t, user2)
	accessToken1 := token1.AccessToken
	accessToken2 := token2.AccessToken

	// 創建 posts（通過 API）
	// User1 創建 post
	createReq1 := map[string]interface{}{
		"content": "User1 Post",
	}
	createResp1 := makeHTTPRequest(t, router, "POST", "/api/v1/posts", createReq1, accessToken1)
	assert.Equal(t, 201, createResp1.Code)

	// User2 創建 post
	createReq2 := map[string]interface{}{
		"content": "User2 Post",
	}
	createResp2 := makeHTTPRequest(t, router, "POST", "/api/v1/posts", createReq2, accessToken2)
	assert.Equal(t, 201, createResp2.Code)

	// 測試按作者過濾
	getResp := makeHTTPRequest(t, router, "GET", fmt.Sprintf("/api/v1/posts?limit=10&author_id=%s", user1.ID), nil, "")
	assert.Equal(t, 200, getResp.Code)

	var response model.CursorResponse[model.Post]
	parseJSONResponse(t, getResp, &response)
	assert.Len(t, response.Data, 1)
	assert.Equal(t, "User1 Post", response.Data[0].Content)
	assert.Equal(t, user1.ID, response.Data[0].AuthorID)
}
