package apihandler

import (
	"fmt"
	"net/http"
)

// MessageKey 消息键类型
type MessageKey string

// 预定义的消息键
const (
	MsgBindError                MessageKey = "bind_error"
	MsgBindErrorDetail          MessageKey = "bind_error_detail"
	MsgPathBindError            MessageKey = "path_bind_error"
	MsgFieldValidationFailed    MessageKey = "field_validation_failed"
	MsgFieldValidationFailedWithParam MessageKey = "field_validation_failed_with_param"
	MsgFieldParseFailed         MessageKey = "field_parse_failed"
	MsgFieldTypeNotSupported    MessageKey = "field_type_not_supported"
)

// Translator 翻译器接口
type Translator interface {
	// Translate 翻译消息
	Translate(key MessageKey, args ...interface{}) string
}

// LocaleFunc 从请求中获取语言环境的函数
type LocaleFunc func(r *http.Request) string

// defaultMessages 默认消息（中文）
var defaultMessages = map[MessageKey]string{
	MsgBindError:                "参数绑定失败",
	MsgBindErrorDetail:          "参数绑定失败: %v",
	MsgPathBindError:            "路径参数绑定失败: %v",
	MsgFieldValidationFailed:    "字段验证失败: %s",
	MsgFieldValidationFailedWithParam: "字段验证失败: %s=%s",
	MsgFieldParseFailed:         "字段 %s 解析失败: %v",
	MsgFieldTypeNotSupported:    "字段 %s 的类型 %s 不支持路径绑定",
}

// englishMessages 英文消息
var englishMessages = map[MessageKey]string{
	MsgBindError:                "Parameter binding failed",
	MsgBindErrorDetail:          "Parameter binding failed: %v",
	MsgPathBindError:            "Path parameter binding failed: %v",
	MsgFieldValidationFailed:    "Field validation failed: %s",
	MsgFieldValidationFailedWithParam: "Field validation failed: %s=%s",
	MsgFieldParseFailed:         "Field %s parsing failed: %v",
	MsgFieldTypeNotSupported:    "Field %s type %s does not support path binding",
}

// SimpleTranslator 简单翻译器实现
type SimpleTranslator struct {
	locale   string
	messages map[MessageKey]string
}

// NewSimpleTranslator 创建简单翻译器
func NewSimpleTranslator(locale string) Translator {
	var messages map[MessageKey]string
	switch locale {
	case "en", "en-US", "en_US":
		messages = englishMessages
	default:
		// 默认使用中文
		messages = defaultMessages
	}
	return &SimpleTranslator{
		locale:   locale,
		messages: messages,
	}
}

// Translate 实现翻译
func (t *SimpleTranslator) Translate(key MessageKey, args ...interface{}) string {
	format, ok := t.messages[key]
	if !ok {
		// 如果找不到翻译，使用默认消息
		format = defaultMessages[key]
	}
	
	if len(args) > 0 {
		return fmt.Sprintf(format, args...)
	}
	return format
}

// DefaultTranslator 默认翻译器（中文）
var DefaultTranslator = NewSimpleTranslator("zh")

// DefaultLocaleFunc 默认语言环境函数（从 Accept-Language 头获取）
var DefaultLocaleFunc = func(r *http.Request) string {
	locale := r.Header.Get("Accept-Language")
	if locale == "" {
		return "zh"
	}
	// 简单处理，只取第一个语言代码
	if len(locale) >= 2 {
		return locale[:2]
	}
	return "zh"
}
