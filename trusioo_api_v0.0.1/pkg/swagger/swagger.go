// Package swagger 提供API文档自动生成功能
package swagger

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Config Swagger配置
type Config struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Version     string            `json:"version"`
	Host        string            `json:"host"`
	BasePath    string            `json:"basePath"`
	Schemes     []string          `json:"schemes"`
	Contact     *Contact          `json:"contact,omitempty"`
	License     *License          `json:"license,omitempty"`
	Extensions  map[string]string `json:"extensions,omitempty"`
}

// Contact 联系信息
type Contact struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

// License 许可证信息
type License struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// SwaggerInfo Swagger文档信息
type SwaggerInfo struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Host        string `json:"host"`
	BasePath    string `json:"basePath"`
}

// APIDefinition API定义
type APIDefinition struct {
	Path        string                `json:"path"`
	Method      string                `json:"method"`
	Summary     string                `json:"summary"`
	Description string                `json:"description"`
	Tags        []string              `json:"tags"`
	Parameters  []Parameter           `json:"parameters"`
	Responses   map[string]Response   `json:"responses"`
	Security    []map[string][]string `json:"security,omitempty"`
}

// Parameter 参数定义
type Parameter struct {
	Name        string      `json:"name"`
	In          string      `json:"in"` // query, header, path, formData, body
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Type        string      `json:"type"`
	Format      string      `json:"format,omitempty"`
	Schema      interface{} `json:"schema,omitempty"`
}

// Response 响应定义
type Response struct {
	Description string      `json:"description"`
	Schema      interface{} `json:"schema,omitempty"`
	Examples    interface{} `json:"examples,omitempty"`
}

// SwaggerDoc Swagger文档生成器
type SwaggerDoc struct {
	config *Config
	apis   []APIDefinition
}

// NewSwaggerDoc 创建新的Swagger文档生成器
func NewSwaggerDoc(config *Config) *SwaggerDoc {
	if config == nil {
		config = &Config{
			Title:       "Trusioo API",
			Description: "企业级API文档",
			Version:     "1.0.0",
			Host:        "localhost:8080",
			BasePath:    "/api/v1",
			Schemes:     []string{"http", "https"},
		}
	}
	return &SwaggerDoc{
		config: config,
		apis:   make([]APIDefinition, 0),
	}
}

// AddAPI 添加API定义
func (s *SwaggerDoc) AddAPI(api APIDefinition) {
	s.apis = append(s.apis, api)
}

// GenerateSpec 生成OpenAPI规范
func (s *SwaggerDoc) GenerateSpec() map[string]interface{} {
	spec := map[string]interface{}{
		"swagger": "2.0",
		"info": map[string]interface{}{
			"title":       s.config.Title,
			"description": s.config.Description,
			"version":     s.config.Version,
		},
		"host":     s.config.Host,
		"basePath": s.config.BasePath,
		"schemes":  s.config.Schemes,
		"consumes": []string{"application/json"},
		"produces": []string{"application/json"},
		"securityDefinitions": map[string]interface{}{
			"BearerAuth": map[string]interface{}{
				"type":        "apiKey",
				"name":        "Authorization",
				"in":          "header",
				"description": "JWT token in the format: Bearer {token}",
			},
		},
		"paths":       s.generatePaths(),
		"definitions": s.generateDefinitions(),
	}

	if s.config.Contact != nil {
		spec["info"].(map[string]interface{})["contact"] = s.config.Contact
	}

	if s.config.License != nil {
		spec["info"].(map[string]interface{})["license"] = s.config.License
	}

	return spec
}

// generatePaths 生成路径定义
func (s *SwaggerDoc) generatePaths() map[string]interface{} {
	paths := make(map[string]interface{})

	for _, api := range s.apis {
		if paths[api.Path] == nil {
			paths[api.Path] = make(map[string]interface{})
		}

		pathItem := paths[api.Path].(map[string]interface{})
		pathItem[strings.ToLower(api.Method)] = map[string]interface{}{
			"summary":     api.Summary,
			"description": api.Description,
			"tags":        api.Tags,
			"parameters":  api.Parameters,
			"responses":   api.Responses,
			"security":    api.Security,
		}
	}

	return paths
}

// generateDefinitions 生成模型定义
func (s *SwaggerDoc) generateDefinitions() map[string]interface{} {
	definitions := map[string]interface{}{
		"Response": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"success": map[string]interface{}{
					"type":        "boolean",
					"description": "请求是否成功",
				},
				"message": map[string]interface{}{
					"type":        "string",
					"description": "响应消息",
				},
				"data": map[string]interface{}{
					"type":        "object",
					"description": "响应数据",
				},
				"error_code": map[string]interface{}{
					"type":        "string",
					"description": "错误码",
				},
				"timestamp": map[string]interface{}{
					"type":        "string",
					"format":      "date-time",
					"description": "响应时间戳",
				},
			},
		},
		"PaginatedResponse": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"success": map[string]interface{}{
					"type": "boolean",
				},
				"message": map[string]interface{}{
					"type": "string",
				},
				"data": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "object",
					},
				},
				"pagination": map[string]interface{}{
					"$ref": "#/definitions/Pagination",
				},
			},
		},
		"Pagination": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"page": map[string]interface{}{
					"type": "integer",
				},
				"limit": map[string]interface{}{
					"type": "integer",
				},
				"total": map[string]interface{}{
					"type": "integer",
				},
				"total_pages": map[string]interface{}{
					"type": "integer",
				},
			},
		},
		"Error": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"success": map[string]interface{}{
					"type":    "boolean",
					"example": false,
				},
				"message": map[string]interface{}{
					"type": "string",
				},
				"error_code": map[string]interface{}{
					"type": "string",
				},
				"details": map[string]interface{}{
					"type": "object",
				},
			},
		},
	}

	return definitions
}

// SetupSwaggerRoutes 设置Swagger路由
func (s *SwaggerDoc) SetupSwaggerRoutes(router *gin.Engine) {
	// Swagger JSON规范端点
	router.GET("/swagger/doc.json", func(c *gin.Context) {
		spec := s.generateSpecWithExamples()
		c.JSON(http.StatusOK, spec)
	})

	// Swagger UI端点
	router.GET("/swagger/*any", func(c *gin.Context) {
		if c.Param("any") == "/" || c.Param("any") == "/index.html" {
			c.Header("Content-Type", "text/html")
			c.String(http.StatusOK, s.generateSwaggerUI())
		} else {
			c.String(http.StatusNotFound, "File not found")
		}
	})
}

// generateSpecWithExamples 生成包含示例的规范
func (s *SwaggerDoc) generateSpecWithExamples() map[string]interface{} {
	spec := s.GenerateSpec()

	// 添加认证相关的API示例
	authAPIs := []APIDefinition{
		{
			Path:        "/admin/login",
			Method:      "POST",
			Summary:     "管理员登录",
			Description: "管理员账户登录接口",
			Tags:        []string{"认证"},
			Parameters: []Parameter{
				{
					Name:        "body",
					In:          "body",
					Description: "登录信息",
					Required:    true,
					Schema: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"email": map[string]interface{}{
								"type":    "string",
								"example": "admin@trusioo.com",
							},
							"password": map[string]interface{}{
								"type":    "string",
								"example": "admin123",
							},
						},
					},
				},
			},
			Responses: map[string]Response{
				"200": {
					Description: "登录成功",
					Schema: map[string]interface{}{
						"$ref": "#/definitions/Response",
					},
				},
				"401": {
					Description: "认证失败",
					Schema: map[string]interface{}{
						"$ref": "#/definitions/Error",
					},
				},
			},
		},
		{
			Path:        "/admin/profile",
			Method:      "GET",
			Summary:     "获取管理员资料",
			Description: "获取当前登录管理员的详细资料",
			Tags:        []string{"认证"},
			Security: []map[string][]string{
				{"BearerAuth": {}},
			},
			Responses: map[string]Response{
				"200": {
					Description: "获取成功",
					Schema: map[string]interface{}{
						"$ref": "#/definitions/Response",
					},
				},
				"401": {
					Description: "未授权",
					Schema: map[string]interface{}{
						"$ref": "#/definitions/Error",
					},
				},
			},
		},
	}

	// 将示例API添加到规范中
	paths := spec["paths"].(map[string]interface{})
	for _, api := range authAPIs {
		if paths[api.Path] == nil {
			paths[api.Path] = make(map[string]interface{})
		}
		pathItem := paths[api.Path].(map[string]interface{})
		pathItem[strings.ToLower(api.Method)] = map[string]interface{}{
			"summary":     api.Summary,
			"description": api.Description,
			"tags":        api.Tags,
			"parameters":  api.Parameters,
			"responses":   api.Responses,
			"security":    api.Security,
		}
	}

	return spec
}

// generateSwaggerUI 生成Swagger UI HTML
func (s *SwaggerDoc) generateSwaggerUI() string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>%s - API文档</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@3.52.5/swagger-ui.css" />
    <style>
        html {
            box-sizing: border-box;
            overflow: -moz-scrollbars-vertical;
            overflow-y: scroll;
        }
        *, *:before, *:after {
            box-sizing: inherit;
        }
        body {
            margin:0;
            background: #fafafa;
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@3.52.5/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@3.52.5/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: '/swagger/doc.json',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout"
            });
        };
    </script>
</body>
</html>
`, s.config.Title)
}

// DefaultConfig 返回默认的Swagger配置
func DefaultConfig() *Config {
	return &Config{
		Title:       "Trusioo API",
		Description: "企业级电商平台API文档 - 提供完整的认证、用户管理、商品管理等功能",
		Version:     "v0.0.1",
		Host:        "localhost:8080",
		BasePath:    "/api/v1",
		Schemes:     []string{"http", "https"},
		Contact: &Contact{
			Name:  "Trusioo团队",
			Email: "dev@trusioo.com",
		},
		License: &License{
			Name: "MIT",
			URL:  "https://opensource.org/licenses/MIT",
		},
	}
}
