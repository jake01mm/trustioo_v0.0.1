package validator

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"trusioo_api_v0.0.1/pkg/errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
)

// ValidatorEngine 验证器引擎
type ValidatorEngine struct {
	validator  *validator.Validate
	translator ut.Translator
}

// ValidationError 验证错误详情
type ValidationError struct {
	Field   string `json:"field"`   // 字段名
	Tag     string `json:"tag"`     // 验证标签
	Message string `json:"message"` // 错误消息
	Value   string `json:"value"`   // 字段值
}

// NewValidator 创建新的验证器
func NewValidator() *ValidatorEngine {
	v := validator.New()

	// 创建中文翻译器
	zhLocale := zh.New()
	uni := ut.New(zhLocale, zhLocale)
	trans, _ := uni.GetTranslator("zh")

	// 注册中文翻译
	zh_translations.RegisterDefaultTranslations(v, trans)

	// 注册自定义验证器
	registerCustomValidators(v, trans)

	// 注册字段名翻译
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		// 如果有中文标签，使用中文名
		if label := fld.Tag.Get("label"); label != "" {
			return label
		}
		return name
	})

	return &ValidatorEngine{
		validator:  v,
		translator: trans,
	}
}

// Validate 验证结构体
func (ve *ValidatorEngine) Validate(obj interface{}) error {
	if err := ve.validator.Struct(obj); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			return ve.formatValidationErrors(validationErrors)
		}
		return errors.NewValidationError(err.Error())
	}
	return nil
}

// formatValidationErrors 格式化验证错误
func (ve *ValidatorEngine) formatValidationErrors(validationErrors validator.ValidationErrors) error {
	var errorDetails []ValidationError

	for _, err := range validationErrors {
		errorDetails = append(errorDetails, ValidationError{
			Field:   err.Field(),
			Tag:     err.Tag(),
			Message: err.Translate(ve.translator),
			Value:   fmt.Sprintf("%v", err.Value()),
		})
	}

	return errors.NewValidationError(errorDetails)
}

// BindAndValidate 绑定并验证请求数据
func (ve *ValidatorEngine) BindAndValidate(c *gin.Context, obj interface{}) error {
	// 绑定请求数据
	if err := c.ShouldBind(obj); err != nil {
		return errors.NewValidationError(fmt.Sprintf("数据绑定失败: %s", err.Error()))
	}

	// 验证数据
	return ve.Validate(obj)
}

// BindJSONAndValidate 绑定JSON并验证
func (ve *ValidatorEngine) BindJSONAndValidate(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		return errors.NewValidationError(fmt.Sprintf("JSON数据绑定失败: %s", err.Error()))
	}

	return ve.Validate(obj)
}

// BindQueryAndValidate 绑定查询参数并验证
func (ve *ValidatorEngine) BindQueryAndValidate(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindQuery(obj); err != nil {
		return errors.NewValidationError(fmt.Sprintf("查询参数绑定失败: %s", err.Error()))
	}

	return ve.Validate(obj)
}

// registerCustomValidators 注册自定义验证器
func registerCustomValidators(v *validator.Validate, trans ut.Translator) {
	// 手机号验证
	v.RegisterValidation("mobile", validateMobile)
	v.RegisterTranslation("mobile", trans, func(ut ut.Translator) error {
		return ut.Add("mobile", "{0} 必须是有效的手机号码", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("mobile", fe.Field())
		return t
	})

	// 强密码验证
	v.RegisterValidation("strong_password", validateStrongPassword)
	v.RegisterTranslation("strong_password", trans, func(ut ut.Translator) error {
		return ut.Add("strong_password", "{0} 必须包含大小写字母、数字和特殊字符，长度至少8位", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("strong_password", fe.Field())
		return t
	})

	// 用户名验证
	v.RegisterValidation("username", validateUsername)
	v.RegisterTranslation("username", trans, func(ut ut.Translator) error {
		return ut.Add("username", "{0} 只能包含字母、数字和下划线，长度3-20位", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("username", fe.Field())
		return t
	})
}

// validateMobile 验证手机号
func validateMobile(fl validator.FieldLevel) bool {
	mobile := fl.Field().String()
	matched, _ := regexp.MatchString(`^1[3-9]\d{9}$`, mobile)
	return matched
}

// validateStrongPassword 验证强密码
func validateStrongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	// 至少8位
	if len(password) < 8 {
		return false
	}

	// 包含大写字母
	hasUpper, _ := regexp.MatchString(`[A-Z]`, password)
	// 包含小写字母
	hasLower, _ := regexp.MatchString(`[a-z]`, password)
	// 包含数字
	hasNumber, _ := regexp.MatchString(`\d`, password)
	// 包含特殊字符
	hasSpecial, _ := regexp.MatchString(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`, password)

	return hasUpper && hasLower && hasNumber && hasSpecial
}

// validateUsername 验证用户名
func validateUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]{3,20}$`, username)
	return matched
}

// 全局验证器实例
var defaultValidator *ValidatorEngine

// Init 初始化全局验证器
func Init() {
	defaultValidator = NewValidator()
}

// GetValidator 获取全局验证器
func GetValidator() *ValidatorEngine {
	if defaultValidator == nil {
		Init()
	}
	return defaultValidator
}

// 便捷函数

// Validate 使用全局验证器验证
func Validate(obj interface{}) error {
	return GetValidator().Validate(obj)
}

// BindAndValidate 使用全局验证器绑定并验证
func BindAndValidate(c *gin.Context, obj interface{}) error {
	return GetValidator().BindAndValidate(c, obj)
}

// BindJSONAndValidate 使用全局验证器绑定JSON并验证
func BindJSONAndValidate(c *gin.Context, obj interface{}) error {
	return GetValidator().BindJSONAndValidate(c, obj)
}

// BindQueryAndValidate 使用全局验证器绑定查询参数并验证
func BindQueryAndValidate(c *gin.Context, obj interface{}) error {
	return GetValidator().BindQueryAndValidate(c, obj)
}
