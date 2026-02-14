package apihandler

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

const (
	// PathTag 路径参数的 tag 名称
	PathTag = "path"
)

// RequestLogger 请求日志记录函数类型
type RequestLogger func(r *http.Request, req any)

// HandleFunc 通用处理函数类型
type HandleFunc[T any, R any] func(ctx context.Context, req *T) (*R, error)

// BizError 业务错误接口
type BizError interface {
	error
	// Code 返回业务错误码
	Code() any
	// HTTPCode 返回 HTTP 状态码
	HTTPCode() int
	// Errors 返回详细错误列表（可选）
	Errors() []any
}

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Code    any    `json:"code"`
	Message string `json:"message"`
	Errors  []any  `json:"errors,omitempty"`
}

// SuccessResponse 成功响应结构
type SuccessResponse[R any] struct {
	Code any `json:"code"`
	Data *R  `json:"data"`
}

// HandlerConfig 处理器配置
type HandlerConfig struct {
	SuccessCode     any
	SuccessHTTPCode int
	BindErrorCode   any
	RequestLogger   RequestLogger // 请求日志记录函数
	Translator      Translator    // 翻译器
	LocaleFunc      LocaleFunc    // 语言环境函数
}

// DefaultConfig 默认配置
var DefaultConfig = &HandlerConfig{
	SuccessCode:     0,
	SuccessHTTPCode: http.StatusOK,
	BindErrorCode:   http.StatusBadRequest,
	RequestLogger:   nil,        // 默认不记录
	Translator:      nil,        // 默认使用中文
	LocaleFunc:      nil,        // 默认使用 Accept-Language
}

// Option 处理器选项函数
type Option func(*HandlerConfig)

// WithSuccessCode 设置成功响应的业务代码
func WithSuccessCode(code any) Option {
	return func(c *HandlerConfig) {
		c.SuccessCode = code
	}
}

// WithSuccessHTTPCode 设置成功响应的 HTTP 状态码
func WithSuccessHTTPCode(code int) Option {
	return func(c *HandlerConfig) {
		c.SuccessHTTPCode = code
	}
}

// WithBindErrorCode 设置参数绑定错误的业务代码
func WithBindErrorCode(code any) Option {
	return func(c *HandlerConfig) {
		c.BindErrorCode = code
	}
}

// WithRequestLogger 设置请求日志记录函数
func WithRequestLogger(logger RequestLogger) Option {
	return func(c *HandlerConfig) {
		c.RequestLogger = logger
	}
}

// WithTranslator 设置翻译器
func WithTranslator(translator Translator) Option {
	return func(c *HandlerConfig) {
		c.Translator = translator
	}
}

// WithLocaleFunc 设置语言环境函数
func WithLocaleFunc(localeFunc LocaleFunc) Option {
	return func(c *HandlerConfig) {
		c.LocaleFunc = localeFunc
	}
}

// Handler 创建 Gin 处理器
func Handler[T any, R any](handleFunc HandleFunc[T, R], opts ...Option) gin.HandlerFunc {
	config := &HandlerConfig{
		SuccessCode:     DefaultConfig.SuccessCode,
		SuccessHTTPCode: DefaultConfig.SuccessHTTPCode,
		BindErrorCode:   DefaultConfig.BindErrorCode,
		RequestLogger:   DefaultConfig.RequestLogger,
		Translator:      DefaultConfig.Translator,
		LocaleFunc:      DefaultConfig.LocaleFunc,
	}
	for _, opt := range opts {
		opt(config)
	}
	return HandlerWithConfig(handleFunc, config)
}

// extractValidationErrors 从验证错误中提取详细信息
func extractValidationErrors(err error, translator Translator) []any {
	var details []any
	
	// 检查是否为验证错误
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			var message string
			// 对于有参数的验证标签，添加参数信息
			if e.Param() != "" {
				message = translator.Translate(MsgFieldValidationFailedWithParam, e.Tag(), e.Param())
			} else {
				message = translator.Translate(MsgFieldValidationFailed, e.Tag())
			}
			details = append(details, map[string]string{
				"field":   e.Field(),
				"message": message,
			})
		}
	}
	
	return details
}

// HandlerWithConfig 使用指定配置创建 Gin 处理器
func HandlerWithConfig[T any, R any](handleFunc HandleFunc[T, R], config *HandlerConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 创建请求对象
		req := new(T)

		// 获取翻译器
		translator := config.Translator
		if translator == nil {
			// 如果未设置翻译器，根据请求获取语言环境
			locale := "zh"
			if config.LocaleFunc != nil {
				locale = config.LocaleFunc(c.Request)
			} else if DefaultLocaleFunc != nil {
				locale = DefaultLocaleFunc(c.Request)
			}
			translator = NewSimpleTranslator(locale)
		}

		// 绑定 JSON/Query 参数
		if err := c.ShouldBind(req); err != nil {
			// 提取验证错误详情
			details := extractValidationErrors(err, translator)
			if len(details) > 0 {
				handleError(c, NewBizErrorWithDetails(config.BindErrorCode, translator.Translate(MsgBindError), http.StatusBadRequest, details))
			} else {
				handleError(c, NewBizError(config.BindErrorCode, translator.Translate(MsgBindErrorDetail, err), http.StatusBadRequest))
			}
			return
		}

		// 绑定路径参数
		if err := bindPathParams(c, req, translator); err != nil {
			handleError(c, NewBizError(config.BindErrorCode, translator.Translate(MsgPathBindError, err), http.StatusBadRequest))
			return
		}

		// 记录请求日志（如果配置了日志函数）
		if config.RequestLogger != nil {
			config.RequestLogger(c.Request, req)
		}

		// 调用业务处理函数
		resp, err := handleFunc(c.Request.Context(), req)
		if err != nil {
			handleError(c, err)
			return
		}

		// 返回成功响应
		c.JSON(config.SuccessHTTPCode, SuccessResponse[R]{
			Code: config.SuccessCode,
			Data: resp,
		})
	}
}

// HandlerWithCode 创建 Gin 处理器，可指定成功响应的 code、HTTP 状态码和参数绑定错误的 code
func HandlerWithCode[T any, R any](handleFunc HandleFunc[T, R], successCode any, successHTTPCode int, bindErrorCode any, requestLogger RequestLogger) gin.HandlerFunc {
	config := &HandlerConfig{
		SuccessCode:     successCode,
		SuccessHTTPCode: successHTTPCode,
		BindErrorCode:   bindErrorCode,
		RequestLogger:   requestLogger,
		Translator:      nil,
		LocaleFunc:      nil,
	}
	return HandlerWithConfig(handleFunc, config)
}

// bindPathParams 绑定路径参数
func bindPathParams(c *gin.Context, req any, translator Translator) error {
	reqType := reflect.TypeOf(req).Elem()
	reqValue := reflect.ValueOf(req).Elem()

	for i := 0; i < reqType.NumField(); i++ {
		field := reqType.Field(i)
		pathTag := field.Tag.Get(PathTag)
		if pathTag == "" {
			continue
		}

		// 从路径中获取参数值
		paramValue := c.Param(pathTag)
		if paramValue == "" {
			continue
		}

		// 根据字段类型进行转换
		fieldValue := reqValue.Field(i)
		if !fieldValue.CanSet() {
			continue
		}

		switch field.Type.Kind() {
		case reflect.String:
			fieldValue.SetString(paramValue)
		case reflect.Int64:
			val, err := strconv.ParseInt(paramValue, 10, 64)
			if err != nil {
				return errors.New(translator.Translate(MsgFieldParseFailed, field.Name, err))
			}
			fieldValue.SetInt(val)
		case reflect.Uint64:
			val, err := strconv.ParseUint(paramValue, 10, 64)
			if err != nil {
				return errors.New(translator.Translate(MsgFieldParseFailed, field.Name, err))
			}
			fieldValue.SetUint(val)
		default:
			return errors.New(translator.Translate(MsgFieldTypeNotSupported, field.Name, field.Type.Kind()))
		}
	}
	return nil
}

// handleError 处理错误
func handleError(c *gin.Context, err error) {
	// 检查是否是业务错误
	if bizErr, ok := err.(BizError); ok {
		c.JSON(bizErr.HTTPCode(), ErrorResponse{
			Code:    bizErr.Code(),
			Message: bizErr.Error(),
			Errors:  bizErr.Errors(),
		})
		return
	}

	// 默认内部服务器错误
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Code:    http.StatusInternalServerError,
		Message: err.Error(),
	})
}
