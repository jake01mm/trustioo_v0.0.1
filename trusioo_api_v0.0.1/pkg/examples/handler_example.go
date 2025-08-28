package examples

import (
	"context"
	"time"

	"trusioo_api_v0.0.1/pkg/errors"
	"trusioo_api_v0.0.1/pkg/logger"
	"trusioo_api_v0.0.1/pkg/response"
	"trusioo_api_v0.0.1/pkg/validator"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ExampleHandler 演示如何使用新工具包的处理器
type ExampleHandler struct {
	logger *logger.Logger
}

// NewExampleHandler 创建演示处理器
func NewExampleHandler(log *logger.Logger) *ExampleHandler {
	return &ExampleHandler{
		logger: log,
	}
}

// UserCreateRequest 用户创建请求
type UserCreateRequest struct {
	Username string `json:"username" validate:"required,username" label:"用户名"`
	Email    string `json:"email" validate:"required,email" label:"邮箱"`
	Password string `json:"password" validate:"required,strong_password" label:"密码"`
	Phone    string `json:"phone" validate:"required,mobile" label:"手机号"`
	Age      int    `json:"age" validate:"required,min=18,max=100" label:"年龄"`
}

// UserResponse 用户响应
type UserResponse struct {
	ID       string    `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Phone    string    `json:"phone"`
	Age      int       `json:"age"`
	CreateAt time.Time `json:"created_at"`
}

// CreateUser 创建用户 - 演示统一响应格式和验证
func (h *ExampleHandler) CreateUser(c *gin.Context) {
	start := time.Now()

	// 使用增强日志记录请求开始
	h.logger.WithRequestContext(c).Info("开始创建用户")

	var req UserCreateRequest

	// 使用新的验证器进行数据绑定和验证
	if err := validator.BindJSONAndValidate(c, &req); err != nil {
		// 记录验证失败
		h.logger.WithRequestContext(c).WithFields(logrus.Fields{
			"error": err.Error(),
		}).Warn("用户创建请求验证失败")

		// 使用统一错误响应
		if appErr, ok := errors.GetAppError(err); ok {
			response.ValidationFailed(c, appErr.Details)
		} else {
			response.BadRequest(c, err.Error())
		}
		return
	}

	// 模拟业务逻辑
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// 检查用户是否已存在（模拟）
	if req.Email == "admin@example.com" {
		// 使用自定义错误
		err := errors.NewDuplicateError("用户")
		h.logger.WithRequestContext(c).WithFields(logrus.Fields{
			"email": req.Email,
		}).Warn("用户已存在")

		response.BusinessError(c, err.Code, err)
		return
	}

	// 模拟数据库操作
	h.simulateCreateUser(ctx, &req)

	// 创建响应数据
	userResponse := &UserResponse{
		ID:       "user-123456",
		Username: req.Username,
		Email:    req.Email,
		Phone:    req.Phone,
		Age:      req.Age,
		CreateAt: time.Now(),
	}

	// 记录性能日志
	duration := time.Since(start)
	h.logger.Performance("create_user", duration, logger.Fields{
		"username": req.Username,
		"email":    req.Email,
	})

	// 记录业务日志
	h.logger.Business("user_created", "success", logger.Fields{
		"user_id":  userResponse.ID,
		"username": req.Username,
		"email":    req.Email,
	})

	// 使用统一成功响应
	response.SuccessWithMessage(c, "用户创建成功", userResponse)
}

// GetUserList 获取用户列表 - 演示分页响应
func (h *ExampleHandler) GetUserList(c *gin.Context) {
	// 查询参数结构
	type QueryParams struct {
		Page   int    `form:"page" validate:"min=1" label:"页码"`
		Limit  int    `form:"limit" validate:"min=1,max=100" label:"每页数量"`
		Search string `form:"search" label:"搜索关键词"`
	}

	var params QueryParams
	params.Page = 1   // 默认值
	params.Limit = 10 // 默认值

	// 绑定查询参数并验证
	if err := validator.BindQueryAndValidate(c, &params); err != nil {
		h.logger.WithRequestContext(c).WithFields(logrus.Fields{
			"error": err.Error(),
		}).Warn("查询参数验证失败")

		response.ValidationFailed(c, err)
		return
	}

	h.logger.WithRequestContext(c).WithFields(logrus.Fields{
		"page":   params.Page,
		"limit":  params.Limit,
		"search": params.Search,
	}).Info("获取用户列表")

	// 模拟查询数据
	users := []UserResponse{
		{
			ID:       "user-1",
			Username: "user1",
			Email:    "user1@example.com",
			Phone:    "13800138001",
			Age:      25,
			CreateAt: time.Now(),
		},
		{
			ID:       "user-2",
			Username: "user2",
			Email:    "user2@example.com",
			Phone:    "13800138002",
			Age:      30,
			CreateAt: time.Now(),
		},
	}

	// 模拟总数
	total := int64(25)

	// 使用分页响应
	response.Paginated(c, users, params.Page, params.Limit, total)
}

// GetUser 获取用户详情 - 演示错误处理
func (h *ExampleHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")

	if userID == "" {
		response.BadRequest(c, "用户ID不能为空")
		return
	}

	h.logger.WithRequestContext(c).WithFields(logrus.Fields{
		"user_id": userID,
	}).Info("获取用户详情")

	// 模拟用户不存在的情况
	if userID == "not-found" {
		err := errors.NewNotFoundError("用户")
		h.logger.WithRequestContext(c).WithFields(logrus.Fields{
			"user_id": userID,
		}).Warn("用户不存在")

		response.BusinessError(c, err.Code, err)
		return
	}

	// 模拟权限不足的情况
	if userID == "forbidden" {
		err := errors.NewForbiddenError()
		h.logger.Security("unauthorized_access", "warning", logger.Fields{
			"user_id":    userID,
			"client_ip":  c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		})

		response.BusinessError(c, err.Code, err)
		return
	}

	// 模拟数据库错误
	if userID == "db-error" {
		err := errors.New(errors.ErrCodeDatabaseError, "连接超时")
		h.logger.WithRequestContext(c).WithFields(logrus.Fields{
			"user_id": userID,
			"error":   err.Error(),
		}).Error("数据库查询失败")

		response.InternalError(c, "服务器内部错误")
		return
	}

	// 模拟成功返回
	user := &UserResponse{
		ID:       userID,
		Username: "demo_user",
		Email:    "demo@example.com",
		Phone:    "13800138000",
		Age:      28,
		CreateAt: time.Now(),
	}

	response.Success(c, user)
}

// simulateCreateUser 模拟创建用户
func (h *ExampleHandler) simulateCreateUser(ctx context.Context, req *UserCreateRequest) {
	start := time.Now()

	// 检查上下文是否已取消
	select {
	case <-ctx.Done():
		h.logger.Warn("创建用户操作被取消", logrus.Fields{"error": ctx.Err()})
		return
	default:
	}

	// 模拟数据库操作耗时
	time.Sleep(100 * time.Millisecond)

	// 记录数据库操作日志（简化版）
	query := "INSERT INTO users (username, email, phone, age) VALUES (?, ?, ?, ?)"
	duration := time.Since(start)

	h.logger.Info("数据库操作完成", logrus.Fields{
		"query":    query,
		"duration": duration,
		"username": req.Username,
		"email":    req.Email,
	})
}

// RegisterRoutes 注册演示路由
func RegisterRoutes(router *gin.RouterGroup, handler *ExampleHandler) {
	examples := router.Group("/examples")
	{
		examples.POST("/users", handler.CreateUser)
		examples.GET("/users", handler.GetUserList)
		examples.GET("/users/:id", handler.GetUser)
	}
}
