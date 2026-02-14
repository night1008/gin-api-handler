package apihandler

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	"github.com/gin-gonic/gin"
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
}

// DefaultConfig 默认配置
var DefaultConfig = &HandlerConfig{
	SuccessCode:     0,
	SuccessHTTPCode: http.StatusOK,
	BindErrorCode:   http.StatusBadRequest,
	RequestLogger:   nil, // 默认不记录
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

// Handler 创建 Gin 处理器
func Handler[T any, R any](handleFunc HandleFunc[T, R], opts ...Option) gin.HandlerFunc {
	config := &HandlerConfig{
		SuccessCode:     DefaultConfig.SuccessCode,
		SuccessHTTPCode: DefaultConfig.SuccessHTTPCode,
		BindErrorCode:   DefaultConfig.BindErrorCode,
		RequestLogger:   DefaultConfig.RequestLogger,
	}
	for _, opt := range opts {
		opt(config)
	}
	return HandlerWithConfig(handleFunc, config)
}

// HandlerWithConfig 使用指定配置创建 Gin 处理器
func HandlerWithConfig[T any, R any](handleFunc HandleFunc[T, R], config *HandlerConfig) gin.HandlerFunc {
	return HandlerWithCode(handleFunc, config.SuccessCode, config.SuccessHTTPCode, config.BindErrorCode, config.RequestLogger)
}

// HandlerWithCode 创建 Gin 处理器，可指定成功响应的 code、HTTP 状态码和参数绑定错误的 code
func HandlerWithCode[T any, R any](handleFunc HandleFunc[T, R], successCode any, successHTTPCode int, bindErrorCode any, requestLogger RequestLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 创建请求对象
		req := new(T)

		// 绑定 JSON/Query 参数
		if err := c.ShouldBind(req); err != nil {
			handleError(c, NewBizError(bindErrorCode, fmt.Sprintf("参数绑定失败: %v", err), http.StatusBadRequest))
			return
		}

		// 绑定路径参数
		if err := bindPathParams(c, req); err != nil {
			handleError(c, NewBizError(bindErrorCode, fmt.Sprintf("路径参数绑定失败: %v", err), http.StatusBadRequest))
			return
		}

		// 记录请求日志（如果配置了日志函数）
		if requestLogger != nil {
			requestLogger(c.Request, req)
		}

		// 调用业务处理函数
		resp, err := handleFunc(c.Request.Context(), req)
		if err != nil {
			handleError(c, err)
			return
		}

		// 返回成功响应
		c.JSON(successHTTPCode, SuccessResponse[R]{
			Code: successCode,
			Data: resp,
		})
	}
}

// bindPathParams 绑定路径参数
func bindPathParams(c *gin.Context, req any) error {
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
				return fmt.Errorf("字段 %s 解析失败: %v", field.Name, err)
			}
			fieldValue.SetInt(val)
		case reflect.Uint64:
			val, err := strconv.ParseUint(paramValue, 10, 64)
			if err != nil {
				return fmt.Errorf("字段 %s 解析失败: %v", field.Name, err)
			}
			fieldValue.SetUint(val)
		default:
			return fmt.Errorf("字段 %s 的类型 %s 不支持路径绑定", field.Name, field.Type.Kind())
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
