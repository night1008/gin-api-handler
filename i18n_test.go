package apihandler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// 测试中文翻译（默认）
func TestI18nChinese(t *testing.T) {
	r := gin.New()

	type testReq struct {
		Name string `json:"name" binding:"required"`
		Age  int    `json:"age" binding:"required,min=18"`
	}

	type testResp struct {
		Message string `json:"message"`
	}

	handleFunc := func(ctx context.Context, req *testReq) (*testResp, error) {
		return &testResp{Message: "success"}, nil
	}

	r.POST("/test", Handler(handleFunc))

	// 发送一个缺少字段的请求
	body := []byte(`{}`)
	req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Language", "zh")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 %d, 实际得到 %d", http.StatusBadRequest, w.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	// 验证消息是中文
	if resp.Message != "参数绑定失败" {
		t.Errorf("期望消息为 '参数绑定失败', 实际得到 '%s'", resp.Message)
	}
}

// 测试英文翻译
func TestI18nEnglish(t *testing.T) {
	r := gin.New()

	type testReq struct {
		Name string `json:"name" binding:"required"`
		Age  int    `json:"age" binding:"required,min=18"`
	}

	type testResp struct {
		Message string `json:"message"`
	}

	handleFunc := func(ctx context.Context, req *testReq) (*testResp, error) {
		return &testResp{Message: "success"}, nil
	}

	r.POST("/test", Handler(handleFunc))

	// 发送一个缺少字段的请求
	body := []byte(`{}`)
	req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Language", "en")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 %d, 实际得到 %d", http.StatusBadRequest, w.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	// 验证消息是英文
	if resp.Message != "Parameter binding failed" {
		t.Errorf("期望消息为 'Parameter binding failed', 实际得到 '%s'", resp.Message)
	}
}

// 测试自定义翻译器
func TestI18nCustomTranslator(t *testing.T) {
	r := gin.New()

	type testReq struct {
		Name string `json:"name" binding:"required"`
	}

	type testResp struct {
		Message string `json:"message"`
	}

	handleFunc := func(ctx context.Context, req *testReq) (*testResp, error) {
		return &testResp{Message: "success"}, nil
	}

	// 使用自定义英文翻译器
	translator := NewSimpleTranslator("en")
	r.POST("/test", Handler(handleFunc, WithTranslator(translator)))

	// 发送一个缺少字段的请求
	body := []byte(`{}`)
	req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 %d, 实际得到 %d", http.StatusBadRequest, w.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	// 验证消息是英文
	if resp.Message != "Parameter binding failed" {
		t.Errorf("期望消息为 'Parameter binding failed', 实际得到 '%s'", resp.Message)
	}
}

// 测试验证错误详情的国际化
func TestI18nValidationErrorDetails(t *testing.T) {
	r := gin.New()

	type testReq struct {
		Name string `json:"name" binding:"required"`
		Age  int    `json:"age" binding:"min=18,max=100"`
	}

	type testResp struct {
		Message string `json:"message"`
	}

	handleFunc := func(ctx context.Context, req *testReq) (*testResp, error) {
		return &testResp{Message: "success"}, nil
	}

	// 使用英文翻译器
	translator := NewSimpleTranslator("en")
	r.POST("/test", Handler(handleFunc, WithTranslator(translator)))

	// 发送一个 age < 18 的请求
	body := []byte(`{"name":"John","age":15}`)
	req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 %d, 实际得到 %d", http.StatusBadRequest, w.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	// 验证包含错误详情
	if len(resp.Errors) == 0 {
		t.Errorf("期望包含详细错误信息，但 errors 字段为空")
	}

	// 验证错误详情的字段信息是英文
	for _, e := range resp.Errors {
		if errDetail, ok := e.(map[string]interface{}); ok {
			if message, ok := errDetail["message"].(string); ok {
				// 验证消息包含 "Field validation failed"
				if len(message) > 0 && message[:5] != "Field" {
					t.Errorf("期望英文错误消息以 'Field' 开头, 实际得到 '%s'", message)
				}
			}
		}
	}
}

// 测试自定义 LocaleFunc
func TestI18nCustomLocaleFunc(t *testing.T) {
	r := gin.New()

	type testReq struct {
		Name string `json:"name" binding:"required"`
	}

	type testResp struct {
		Message string `json:"message"`
	}

	handleFunc := func(ctx context.Context, req *testReq) (*testResp, error) {
		return &testResp{Message: "success"}, nil
	}

	// 自定义 locale 函数，从 query 参数获取语言
	customLocaleFunc := func(r *http.Request) string {
		lang := r.URL.Query().Get("lang")
		if lang == "" {
			return "zh"
		}
		return lang
	}

	r.POST("/test", Handler(handleFunc, WithLocaleFunc(customLocaleFunc)))

	// 发送一个缺少字段的请求，并指定英文
	body := []byte(`{}`)
	req := httptest.NewRequest("POST", "/test?lang=en", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 %d, 实际得到 %d", http.StatusBadRequest, w.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	// 验证消息是英文
	if resp.Message != "Parameter binding failed" {
		t.Errorf("期望消息为 'Parameter binding failed', 实际得到 '%s'", resp.Message)
	}
}

// 测试路径参数解析错误的国际化
func TestI18nPathParamError(t *testing.T) {
	r := gin.New()

	type testReq struct {
		ID int64 `path:"id"`
	}

	type testResp struct {
		ID int64 `json:"id"`
	}

	handleFunc := func(ctx context.Context, req *testReq) (*testResp, error) {
		return &testResp{ID: req.ID}, nil
	}

	// 使用英文翻译器
	translator := NewSimpleTranslator("en")
	r.GET("/test/:id", Handler(handleFunc, WithTranslator(translator)))

	// 发送一个无效的 ID（非数字）
	req := httptest.NewRequest("GET", "/test/invalid", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 %d, 实际得到 %d", http.StatusBadRequest, w.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	// 验证消息包含 "Path parameter binding failed"
	if resp.Message[:4] != "Path" {
		t.Errorf("期望消息以 'Path' 开头, 实际得到 '%s'", resp.Message)
	}
}
