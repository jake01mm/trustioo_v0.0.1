package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// AuthMiddleware 认证中间件结构
type AuthMiddleware struct {
	jwtManager *JWTManager
	logger     *logrus.Logger
}

// NewAuthMiddleware 创建新的认证中间件
func NewAuthMiddleware(jwtManager *JWTManager, logger *logrus.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
		logger:     logger,
	}
}

// RequireAuth 要求认证的中间件
func (am *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := am.extractToken(c)
		if err != nil {
			am.logger.WithError(err).Warn("Token extraction failed")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		claims, err := am.jwtManager.ValidateToken(token)
		if err != nil {
			am.logger.WithError(err).Warn("Token validation failed")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid token",
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Set("user_type", claims.UserType)
		c.Set("claims", claims)

		c.Next()
	}
}

// RequireRole 要求特定角色的中间件
func (am *AuthMiddleware) RequireRole(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 首先检查是否已经通过认证
		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Authentication required",
			})
			c.Abort()
			return
		}

		userClaims := claims.(*Claims)

		// 检查用户角色是否在允许的角色列表中
		for _, role := range requiredRoles {
			if userClaims.Role == role {
				c.Next()
				return
			}
		}

		am.logger.WithFields(logrus.Fields{
			"user_id":        userClaims.UserID,
			"user_role":      userClaims.Role,
			"required_roles": requiredRoles,
		}).Warn("Access denied: insufficient role")

		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Forbidden",
			"message": "Insufficient privileges",
		})
		c.Abort()
	}
}

// RequireUserType 要求特定用户类型的中间件
func (am *AuthMiddleware) RequireUserType(requiredTypes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 首先检查是否已经通过认证
		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Authentication required",
			})
			c.Abort()
			return
		}

		userClaims := claims.(*Claims)

		// 检查用户类型是否在允许的类型列表中
		for _, userType := range requiredTypes {
			if userClaims.UserType == userType {
				c.Next()
				return
			}
		}

		am.logger.WithFields(logrus.Fields{
			"user_id":        userClaims.UserID,
			"user_type":      userClaims.UserType,
			"required_types": requiredTypes,
		}).Warn("Access denied: insufficient user type")

		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Forbidden",
			"message": "Access denied for this user type",
		})
		c.Abort()
	}
}

// OptionalAuth 可选认证中间件（不强制要求认证）
func (am *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := am.extractToken(c)
		if err != nil {
			// 没有令牌或令牌格式错误，继续处理但不设置用户信息
			c.Next()
			return
		}

		claims, err := am.jwtManager.ValidateToken(token)
		if err != nil {
			// 令牌无效，记录警告但继续处理
			am.logger.WithError(err).Warn("Invalid token in optional auth")
			c.Next()
			return
		}

		// 设置用户信息到上下文
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Set("user_type", claims.UserType)
		c.Set("claims", claims)

		c.Next()
	}
}

// extractToken 从请求中提取令牌
func (am *AuthMiddleware) extractToken(c *gin.Context) (string, error) {
	// 首先尝试从Authorization头中获取
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		return ExtractTokenFromHeader(authHeader)
	}

	// 尝试从查询参数中获取（不推荐，但作为备选方案）
	token := c.Query("token")
	if token != "" {
		return token, nil
	}

	return "", errors.New("no token provided")
}

// GetCurrentUser 获取当前用户信息
func GetCurrentUser(c *gin.Context) (*Claims, error) {
	claims, exists := c.Get("claims")
	if !exists {
		return nil, errors.New("user not authenticated")
	}

	userClaims, ok := claims.(*Claims)
	if !ok {
		return nil, errors.New("invalid user claims")
	}

	return userClaims, nil
}

// GetCurrentUserID 获取当前用户ID
func GetCurrentUserID(c *gin.Context) (string, error) {
	claims, err := GetCurrentUser(c)
	if err != nil {
		return "", err
	}
	return claims.UserID, nil
}

// GetCurrentUserEmail 获取当前用户邮箱
func GetCurrentUserEmail(c *gin.Context) (string, error) {
	claims, err := GetCurrentUser(c)
	if err != nil {
		return "", err
	}
	return claims.Email, nil
}

// GetCurrentUserRole 获取当前用户角色
func GetCurrentUserRole(c *gin.Context) (string, error) {
	claims, err := GetCurrentUser(c)
	if err != nil {
		return "", err
	}
	return claims.Role, nil
}

// GetCurrentUserType 获取当前用户类型
func GetCurrentUserType(c *gin.Context) (string, error) {
	claims, err := GetCurrentUser(c)
	if err != nil {
		return "", err
	}
	return claims.UserType, nil
}

// IsAuthenticated 检查用户是否已认证
func IsAuthenticated(c *gin.Context) bool {
	_, exists := c.Get("claims")
	return exists
}

// HasRole 检查用户是否具有指定角色
func HasRole(c *gin.Context, role string) bool {
	claims, err := GetCurrentUser(c)
	if err != nil {
		return false
	}
	return claims.Role == role
}

// HasUserType 检查用户是否为指定类型
func HasUserType(c *gin.Context, userType string) bool {
	claims, err := GetCurrentUser(c)
	if err != nil {
		return false
	}
	return claims.UserType == userType
}
