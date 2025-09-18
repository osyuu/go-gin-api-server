package integration

import (
	"bytes"
	"encoding/json"
	"go-gin-api-server/config"
	"go-gin-api-server/internal/database"
	"go-gin-api-server/internal/model"
	"go-gin-api-server/internal/repository"
	"go-gin-api-server/pkg/logger"
	"go-gin-api-server/pkg/utils"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	NonExistentUserID = "550e8400-e29b-41d4-a716-446655440000"
	NonExistentPostID = 999999999999
	InvalidCursorID   = "invalid-id"
	testAdminUserID   = "admin-user-id-6734"
)

// 全局變量，在 TestMain 中初始化
var (
	globalConfig     *config.Config
	globalJWTManager *utils.JWTManager
)

// TestMain 設置測試環境
func TestMain(m *testing.M) {
	// 初始化測試專用配置和日誌
	globalConfig = config.LoadTestConfig()
	logger.Init(config.Test)

	// 確認 TestMain 被執行
	logger.Log.Info("🚀 TestMain started - Initializing integration test environment")

	// 初始化 JWT Manager
	globalJWTManager = utils.NewJWTManager(globalConfig.JWT.Secret, globalConfig.JWT.AccessTokenExpiration)

	// 初始化測試資料庫
	if err := database.InitDatabase(globalConfig.Database); err != nil {
		panic("Failed to initialize test database: " + err.Error())
	}

	// 運行資料庫 migration
	// 注意：Integration 測試應該使用 golang-migrate 來管理資料庫結構
	// 請確保在運行 integration 測試前先執行: make migrate-test-up
	db := database.GetDB()
	// 檢查表是否存在，如果不存在則提示用戶運行 migration
	var count int64
	if err := db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'users'").Scan(&count).Error; err != nil {
		panic("Failed to check database tables: " + err.Error())
	}
	if count == 0 {
		panic("Database tables not found. Please ensure test database is set up with migrations.")
	}

	// 運行測試
	logger.Log.Info("🧪 Running integration tests...")
	code := m.Run()

	// 測試完成
	logger.Log.Info("✅ Integration tests completed", zap.Int("exit_code", code))

	// 退出
	os.Exit(code)
}

// setup 為每個測試準備乾淨的資料庫狀態
func setup() *gorm.DB {
	// 清理測試數據
	db := database.GetDB()
	if db != nil {
		db.Exec("DELETE FROM user_credentials")
		db.Exec("DELETE FROM users")
	}
	return db
}

// teardown 測試後的清理處理
func teardown(db *gorm.DB) {
	// 清理測試數據
	if db != nil {
		db.Exec("DELETE FROM user_credentials")
		db.Exec("DELETE FROM users")
	}
}

// Helper function

func makeHTTPRequest(t *testing.T, router *gin.Engine, method, url string, body interface{}, token string) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBody, err := json.Marshal(body)
		assert.NoError(t, err)
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, url, reqBody)
	assert.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// makeHTTPRequestWithCookie 發送帶有 Cookie 的 HTTP 請求
func makeHTTPRequestWithCookie(t *testing.T, router *gin.Engine, method, url string, body interface{}, refreshToken string) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBody, err := json.Marshal(body)
		assert.NoError(t, err)
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, url, reqBody)
	assert.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	// 添加 refresh token cookie
	if refreshToken != "" {
		req.AddCookie(&http.Cookie{
			Name:  "gin_api_refresh_token",
			Value: refreshToken,
		})
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func parseJSONResponse(t *testing.T, w *httptest.ResponseRecorder, target interface{}) {
	err := json.Unmarshal(w.Body.Bytes(), target)
	assert.NoError(t, err)
}

func createTestUser(t *testing.T, db *gorm.DB, overrides ...map[string]interface{}) *model.User {
	username := "testuser"
	email := "test@example.com"
	birthDate := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)

	if len(overrides) > 0 {
		override := overrides[0]
		if val, ok := override["username"]; ok {
			username = val.(string)
		}
		if val, ok := override["email"]; ok {
			email = val.(string)
		}
		if val, ok := override["birth_date"]; ok {
			birthDate = val.(time.Time)
		}
	}

	user := model.CreateUser(
		"Test User",
		&username,
		&email,
		&birthDate,
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
