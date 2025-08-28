// Package testframework 提供测试框架和工具
package testframework

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"trusioo_api_v0.0.1/internal/config"
	"trusioo_api_v0.0.1/internal/infrastructure/database"
	"trusioo_api_v0.0.1/pkg/cache"
	"trusioo_api_v0.0.1/pkg/response"
)

// TestSuite 基础测试套件
type TestSuite struct {
	suite.Suite
	Router     *gin.Engine
	DB         *database.Database
	Redis      *redis.Client
	Cache      cache.Cache
	Logger     *logrus.Logger
	TestConfig *TestConfig
}

// TestConfig 测试配置
type TestConfig struct {
	DatabaseURL string `json:"database_url"`
	RedisURL    string `json:"redis_url"`
	LogLevel    string `json:"log_level"`
	TestData    string `json:"test_data"`
	CleanupMode string `json:"cleanup_mode"` // auto, manual, skip
}

// APITestCase API测试用例
type APITestCase struct {
	Name           string                 `json:"name"`
	Method         string                 `json:"method"`
	URL            string                 `json:"url"`
	Headers        map[string]string      `json:"headers"`
	Body           interface{}            `json:"body"`
	ExpectedStatus int                    `json:"expected_status"`
	ExpectedBody   map[string]interface{} `json:"expected_body"`
	Setup          func(*TestSuite)       `json:"-"`
	Cleanup        func(*TestSuite)       `json:"-"`
}

// MockData 模拟数据生成器
type MockData struct {
	Users    []MockUser    `json:"users"`
	Admins   []MockAdmin   `json:"admins"`
	Products []MockProduct `json:"products"`
}

// MockUser 模拟用户数据
type MockUser struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
	Active   bool   `json:"active"`
}

// MockAdmin 模拟管理员数据
type MockAdmin struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
	Role     string `json:"role"`
	Active   bool   `json:"active"`
}

// MockProduct 模拟商品数据
type MockProduct struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	Active      bool    `json:"active"`
}

// NewTestSuite 创建新的测试套件
func NewTestSuite(config *TestConfig) *TestSuite {
	if config == nil {
		config = DefaultTestConfig()
	}

	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	// 创建logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	if config.LogLevel != "" {
		if level, err := logrus.ParseLevel(config.LogLevel); err == nil {
			logger.SetLevel(level)
		}
	}

	return &TestSuite{
		TestConfig: config,
		Logger:     logger,
	}
}

// SetupSuite 设置测试套件
func (ts *TestSuite) SetupSuite() {
	ts.Logger.Info("Setting up test suite")

	// 设置数据库连接
	if ts.TestConfig.DatabaseURL != "" {
		ts.setupDatabase()
	}

	// 设置Redis连接
	if ts.TestConfig.RedisURL != "" {
		ts.setupRedis()
	}

	// 设置缓存
	if ts.Redis != nil {
		redisCache := cache.NewRedisCache(ts.Redis, "test", ts.Logger)
		ts.Cache = redisCache
	}

	// 设置路由
	ts.setupRouter()

	// 加载测试数据
	if ts.TestConfig.TestData != "" {
		ts.loadTestData()
	}
}

// TearDownSuite 清理测试套件
func (ts *TestSuite) TearDownSuite() {
	ts.Logger.Info("Tearing down test suite")

	if ts.TestConfig.CleanupMode == "auto" {
		ts.cleanupTestData()
	}

	if ts.DB != nil {
		ts.DB.Close()
	}

	if ts.Redis != nil {
		ts.Redis.Close()
	}
}

// setupDatabase 设置测试数据库
func (ts *TestSuite) setupDatabase() {
	db, err := database.New(&config.DatabaseConfig{
		Host:            "localhost",
		Port:            "5432",
		Name:            "trusioo_api_test",
		User:            "postgres",
		Password:        "password",
		SSLMode:         "disable",
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Hour,
	}, ts.Logger)

	if err != nil {
		ts.T().Fatalf("Failed to connect to test database: %v", err)
	}

	ts.DB = db
}

// setupRedis 设置测试Redis
func (ts *TestSuite) setupRedis() {
	ts.Redis = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       1, // 使用不同的DB以避免与开发环境冲突
	})

	ctx := context.Background()
	if err := ts.Redis.Ping(ctx).Err(); err != nil {
		ts.T().Fatalf("Failed to connect to test Redis: %v", err)
	}
}

// setupRouter 设置测试路由
func (ts *TestSuite) setupRouter() {
	ts.Router = gin.New()

	// 添加基础中间件
	ts.Router.Use(gin.Recovery())

	// 添加测试专用的日志中间件
	ts.Router.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		ts.Logger.WithFields(logrus.Fields{
			"method":   c.Request.Method,
			"path":     c.Request.URL.Path,
			"status":   c.Writer.Status(),
			"duration": duration,
		}).Info("Test request completed")
	})
}

// loadTestData 加载测试数据
func (ts *TestSuite) loadTestData() {
	if _, err := os.Stat(ts.TestConfig.TestData); os.IsNotExist(err) {
		ts.Logger.Warn("Test data file not found, creating default test data")
		ts.createDefaultTestData()
		return
	}

	data, err := os.ReadFile(ts.TestConfig.TestData)
	if err != nil {
		ts.T().Fatalf("Failed to read test data: %v", err)
	}

	var mockData MockData
	if err := json.Unmarshal(data, &mockData); err != nil {
		ts.T().Fatalf("Failed to parse test data: %v", err)
	}

	ts.insertTestData(&mockData)
}

// createDefaultTestData 创建默认测试数据
func (ts *TestSuite) createDefaultTestData() {
	mockData := &MockData{
		Users: []MockUser{
			{
				ID:       "user-test-1",
				Email:    "testuser1@example.com",
				Name:     "Test User 1",
				Password: "password123",
				Active:   true,
			},
			{
				ID:       "user-test-2",
				Email:    "testuser2@example.com",
				Name:     "Test User 2",
				Password: "password123",
				Active:   true,
			},
		},
		Admins: []MockAdmin{
			{
				ID:       "admin-test-1",
				Email:    "testadmin@example.com",
				Name:     "Test Admin",
				Password: "admin123",
				Role:     "admin",
				Active:   true,
			},
		},
		Products: []MockProduct{
			{
				ID:          "product-test-1",
				Name:        "Test Product 1",
				Description: "This is a test product",
				Price:       99.99,
				Stock:       100,
				Active:      true,
			},
		},
	}

	ts.insertTestData(mockData)
}

// insertTestData 插入测试数据
func (ts *TestSuite) insertTestData(mockData *MockData) {
	if ts.DB == nil {
		return
	}

	// 插入测试用户
	for _, user := range mockData.Users {
		query := `
			INSERT INTO users (id, email, name, password, active, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
			ON CONFLICT (id) DO NOTHING
		`
		_, err := ts.DB.Exec(query, user.ID, user.Email, user.Name, user.Password, user.Active)
		if err != nil {
			ts.Logger.WithError(err).Warn("Failed to insert test user")
		}
	}

	// 插入测试管理员
	for _, admin := range mockData.Admins {
		query := `
			INSERT INTO admins (id, email, name, password, role, active, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
			ON CONFLICT (id) DO NOTHING
		`
		_, err := ts.DB.Exec(query, admin.ID, admin.Email, admin.Name, admin.Password, admin.Role, admin.Active)
		if err != nil {
			ts.Logger.WithError(err).Warn("Failed to insert test admin")
		}
	}
}

// cleanupTestData 清理测试数据
func (ts *TestSuite) cleanupTestData() {
	if ts.DB == nil {
		return
	}

	tables := []string{"users", "admins", "products"}
	for _, table := range tables {
		query := fmt.Sprintf("DELETE FROM %s WHERE id LIKE '%%test%%'", table)
		_, err := ts.DB.Exec(query)
		if err != nil {
			ts.Logger.WithError(err).Warnf("Failed to cleanup table %s", table)
		}
	}

	// 清理Redis测试数据
	if ts.Redis != nil {
		ctx := context.Background()
		ts.Redis.FlushDB(ctx)
	}
}

// MakeRequest 发送HTTP请求
func (ts *TestSuite) MakeRequest(method, url string, body interface{}) *httptest.ResponseRecorder {
	var reader io.Reader

	if body != nil {
		jsonBody, err := json.Marshal(body)
		ts.Require().NoError(err)
		reader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, url, reader)
	ts.Require().NoError(err)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	w := httptest.NewRecorder()
	ts.Router.ServeHTTP(w, req)

	return w
}

// MakeRequestWithHeaders 发送带头部的HTTP请求
func (ts *TestSuite) MakeRequestWithHeaders(method, url string, headers map[string]string, body interface{}) *httptest.ResponseRecorder {
	var reader io.Reader

	if body != nil {
		jsonBody, err := json.Marshal(body)
		ts.Require().NoError(err)
		reader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, url, reader)
	ts.Require().NoError(err)

	// 设置头部
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	if body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	w := httptest.NewRecorder()
	ts.Router.ServeHTTP(w, req)

	return w
}

// AssertSuccessResponse 断言成功响应
func (ts *TestSuite) AssertSuccessResponse(w *httptest.ResponseRecorder, expectedStatus int) {
	ts.Assert().Equal(expectedStatus, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	ts.Require().NoError(err)

	ts.Assert().True(resp.Success)
	ts.Assert().NotEmpty(resp.Message)
	ts.Assert().NotEmpty(resp.Timestamp)
}

// AssertErrorResponse 断言错误响应
func (ts *TestSuite) AssertErrorResponse(w *httptest.ResponseRecorder, expectedStatus int, expectedErrorCode int) {
	ts.Assert().Equal(expectedStatus, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	ts.Require().NoError(err)

	ts.Assert().False(resp.Success)
	ts.Assert().NotEmpty(resp.Message)
	ts.Assert().Equal(expectedErrorCode, resp.Code)
	ts.Assert().NotEmpty(resp.Timestamp)
}

// AssertResponseContains 断言响应包含特定内容
func (ts *TestSuite) AssertResponseContains(w *httptest.ResponseRecorder, key string, expectedValue interface{}) {
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	ts.Require().NoError(err)

	if data, ok := resp["data"].(map[string]interface{}); ok {
		ts.Assert().Equal(expectedValue, data[key])
	} else {
		ts.Assert().Equal(expectedValue, resp[key])
	}
}

// RunAPITestCases 运行API测试用例
func (ts *TestSuite) RunAPITestCases(testCases []APITestCase) {
	for _, tc := range testCases {
		ts.T().Run(tc.Name, func(t *testing.T) {
			// 执行setup
			if tc.Setup != nil {
				tc.Setup(ts)
			}

			// 发送请求
			w := ts.MakeRequestWithHeaders(tc.Method, tc.URL, tc.Headers, tc.Body)

			// 检查状态码
			assert.Equal(t, tc.ExpectedStatus, w.Code, "Status code mismatch")

			// 检查响应体
			if tc.ExpectedBody != nil {
				var actualBody map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &actualBody)
				assert.NoError(t, err, "Failed to parse response body")

				for key, expectedValue := range tc.ExpectedBody {
					assert.Equal(t, expectedValue, actualBody[key], "Response body field %s mismatch", key)
				}
			}

			// 执行cleanup
			if tc.Cleanup != nil {
				tc.Cleanup(ts)
			}
		})
	}
}

// DatabaseTransaction 数据库事务测试辅助函数
func (ts *TestSuite) DatabaseTransaction(fn func(*sql.Tx)) {
	tx, err := ts.DB.Begin()
	ts.Require().NoError(err)

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
		tx.Rollback() // 测试中总是回滚
	}()

	fn(tx)
}

// MockTime 模拟时间辅助函数
func (ts *TestSuite) MockTime(mockTime time.Time, fn func()) {
	// 这里可以实现时间模拟逻辑
	// 由于Go的time包不容易mock，这里提供一个框架
	fn()
}

// GenerateJWTToken 生成测试JWT token
func (ts *TestSuite) GenerateJWTToken(userID, userType string) string {
	// 这里应该使用实际的JWT生成逻辑
	// 为了示例，返回一个模拟token
	return fmt.Sprintf("test-jwt-token-%s-%s", userType, userID)
}

// AssertCacheValue 断言缓存值
func (ts *TestSuite) AssertCacheValue(key string, expectedValue interface{}) {
	if ts.Cache == nil {
		ts.T().Skip("Cache not available")
		return
	}

	ctx := context.Background()
	actualBytes, err := ts.Cache.Get(ctx, key)

	if expectedValue == nil {
		ts.Assert().Error(err, "Expected cache miss")
		return
	}

	ts.Require().NoError(err)

	var actualValue interface{}
	err = json.Unmarshal(actualBytes, &actualValue)
	ts.Require().NoError(err)

	ts.Assert().Equal(expectedValue, actualValue)
}

// SetCacheValue 设置缓存值
func (ts *TestSuite) SetCacheValue(key string, value interface{}, ttl time.Duration) {
	if ts.Cache == nil {
		return
	}

	ctx := context.Background()
	err := ts.Cache.Set(ctx, key, value, ttl)
	ts.Require().NoError(err)
}

// DefaultTestConfig 返回默认测试配置
func DefaultTestConfig() *TestConfig {
	return &TestConfig{
		DatabaseURL: "postgres://postgres:password@localhost:5432/trusioo_api_test?sslmode=disable",
		RedisURL:    "redis://localhost:6379/1",
		LogLevel:    "info",
		TestData:    "testdata/mock_data.json",
		CleanupMode: "auto",
	}
}

// NewAPITestCase 创建新的API测试用例
func NewAPITestCase(name, method, url string) *APITestCase {
	return &APITestCase{
		Name:           name,
		Method:         strings.ToUpper(method),
		URL:            url,
		Headers:        make(map[string]string),
		ExpectedStatus: http.StatusOK,
		ExpectedBody:   make(map[string]interface{}),
	}
}

// WithHeaders 添加请求头
func (tc *APITestCase) WithHeaders(headers map[string]string) *APITestCase {
	for k, v := range headers {
		tc.Headers[k] = v
	}
	return tc
}

// WithAuth 添加认证头
func (tc *APITestCase) WithAuth(token string) *APITestCase {
	tc.Headers["Authorization"] = fmt.Sprintf("Bearer %s", token)
	return tc
}

// WithBody 设置请求体
func (tc *APITestCase) WithBody(body interface{}) *APITestCase {
	tc.Body = body
	return tc
}

// ExpectStatus 设置期望状态码
func (tc *APITestCase) ExpectStatus(status int) *APITestCase {
	tc.ExpectedStatus = status
	return tc
}

// ExpectField 设置期望的响应字段
func (tc *APITestCase) ExpectField(key string, value interface{}) *APITestCase {
	tc.ExpectedBody[key] = value
	return tc
}

// WithSetup 设置测试前置条件
func (tc *APITestCase) WithSetup(setup func(*TestSuite)) *APITestCase {
	tc.Setup = setup
	return tc
}

// WithCleanup 设置测试清理
func (tc *APITestCase) WithCleanup(cleanup func(*TestSuite)) *APITestCase {
	tc.Cleanup = cleanup
	return tc
}
