package apihandler

import "net/http"

// BaseBizError 基础业务错误
type BaseBizError struct {
	code     any
	message  string
	httpCode int
	errors   []any
}

// NewBizError 创建业务错误
func NewBizError(code any, message string, httpCode int) BizError {
	return &BaseBizError{
		code:     code,
		message:  message,
		httpCode: httpCode,
	}
}

// NewBizErrorWithDetails 创建带详细错误的业务错误
func NewBizErrorWithDetails(code any, message string, httpCode int, errors []any) BizError {
	return &BaseBizError{
		code:     code,
		message:  message,
		httpCode: httpCode,
		errors:   errors,
	}
}

// Error 实现 error 接口
func (e *BaseBizError) Error() string {
	return e.message
}

// Code 返回业务错误码
func (e *BaseBizError) Code() any {
	return e.code
}

// HTTPCode 返回 HTTP 状态码
func (e *BaseBizError) HTTPCode() int {
	return e.httpCode
}

// Errors 返回详细错误列表
func (e *BaseBizError) Errors() []any {
	return e.errors
}

// 预定义的常见业务错误
var (
	// ErrBadRequest 请求参数错误
	ErrBadRequest = func(code any, msg string) BizError {
		return NewBizError(code, msg, http.StatusBadRequest)
	}

	// ErrUnauthorized 未授权
	ErrUnauthorized = func(code any, msg string) BizError {
		return NewBizError(code, msg, http.StatusUnauthorized)
	}

	// ErrForbidden 禁止访问
	ErrForbidden = func(code any, msg string) BizError {
		return NewBizError(code, msg, http.StatusForbidden)
	}

	// ErrNotFound 资源不存在
	ErrNotFound = func(code any, msg string) BizError {
		return NewBizError(code, msg, http.StatusNotFound)
	}

	// ErrConflict 资源冲突
	ErrConflict = func(code any, msg string) BizError {
		return NewBizError(code, msg, http.StatusConflict)
	}

	// ErrInternalServer 内部服务器错误
	ErrInternalServer = func(code any, msg string) BizError {
		return NewBizError(code, msg, http.StatusInternalServerError)
	}
)
