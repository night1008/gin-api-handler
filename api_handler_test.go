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

func init() {
	gin.SetMode(gin.TestMode)
}

// 测试请求结构
type testRequest struct {
	ID   int64  `path:"id"`
	Name string `form:"name"`
	Age  int    `form:"age"`
}

// 测试响应结构
type testResponse struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Age     int    `json:"age"`
	Message string `json:"message"`
}

// 测试成功场景
func TestHandlerSuccess(t *testing.T) {
	r := gin.New()

	handleFunc := func(ctx context.Context, req *testRequest) (*testResponse, error) {
		return &testResponse{
			ID:      req.ID,
			Name:    req.Name,
			Age:     req.Age,
			Message: "success",
		}, nil
	}

	r.GET("/test/:id", Handler(handleFunc))

	req := httptest.NewRequest("GET", "/test/123?age=25&name=张三", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际得到 %d", http.StatusOK, w.Code)
	}

	var resp SuccessResponse[testResponse]
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	// JSON 解码时数字会被解析为 float64
	code, ok := resp.Code.(float64)
	if !ok {
		t.Fatalf("期望 code 为数字类型，实际得到 %T", resp.Code)
	}

	if int(code) != 0 {
		t.Errorf("期望 code 为 0, 实际得到 %v", code)
	}

	if resp.Data.ID != 123 {
		t.Errorf("期望 ID 为 123, 实际得到 %d", resp.Data.ID)
	}

	if resp.Data.Name != "张三" {
		t.Errorf("期望 Name 为 '张三', 实际得到 '%s'", resp.Data.Name)
	}

	if resp.Data.Age != 25 {
		t.Errorf("期望 Age 为 25, 实际得到 %d", resp.Data.Age)
	}
}

// 测试业务错误
func TestHandlerBizError(t *testing.T) {
	r := gin.New()

	handleFunc := func(ctx context.Context, req *testRequest) (*testResponse, error) {
		if req.ID == 0 {
			return nil, ErrBadRequest(40000, "ID不能为0")
		}
		return nil, ErrNotFound(40400, "资源不存在")
	}

	r.GET("/test/:id", Handler(handleFunc))

	req := httptest.NewRequest("GET", "/test/999", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("期望状态码 %d, 实际得到 %d", http.StatusNotFound, w.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	// JSON 解码时数字会被解析为 float64
	code, ok := resp.Code.(float64)
	if !ok {
		t.Fatalf("期望 code 为数字类型，实际得到 %T", resp.Code)
	}

	if int(code) != 40400 {
		t.Errorf("期望 code 为 40400, 实际得到 %v", code)
	}
}

// 测试路径参数类型 - uint64
func TestHandlerUint64Path(t *testing.T) {
	type uint64Request struct {
		ID uint64 `path:"id"`
	}

	type uint64Response struct {
		ID uint64 `json:"id"`
	}

	r := gin.New()

	handleFunc := func(ctx context.Context, req *uint64Request) (*uint64Response, error) {
		return &uint64Response{ID: req.ID}, nil
	}

	r.GET("/test/:id", Handler(handleFunc))

	req := httptest.NewRequest("GET", "/test/18446744073709551615", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际得到 %d", http.StatusOK, w.Code)
	}
}

// 测试路径参数类型 - string
func TestHandlerStringPath(t *testing.T) {
	type stringRequest struct {
		Name string `path:"name"`
	}

	type stringResponse struct {
		Name string `json:"name"`
	}

	r := gin.New()

	handleFunc := func(ctx context.Context, req *stringRequest) (*stringResponse, error) {
		return &stringResponse{Name: req.Name}, nil
	}

	r.GET("/test/:name", Handler(handleFunc))

	req := httptest.NewRequest("GET", "/test/hello", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际得到 %d", http.StatusOK, w.Code)
	}

	var resp SuccessResponse[stringResponse]
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	if resp.Data.Name != "hello" {
		t.Errorf("期望 Name 为 'hello', 实际得到 '%s'", resp.Data.Name)
	}
}

// 测试自定义业务错误
func TestCustomBizError(t *testing.T) {
	customErr := NewBizError(10001, "自定义错误", http.StatusBadRequest)

	code, ok := customErr.Code().(int)
	if !ok {
		t.Fatalf("期望 code 为 int 类型，实际得到 %T", customErr.Code())
	}

	if code != 10001 {
		t.Errorf("期望 code 为 10001, 实际得到 %d", code)
	}

	if customErr.HTTPCode() != http.StatusBadRequest {
		t.Errorf("期望 HTTP code 为 %d, 实际得到 %d", http.StatusBadRequest, customErr.HTTPCode())
	}

	if customErr.Error() != "自定义错误" {
		t.Errorf("期望错误消息为 '自定义错误', 实际得到 '%s'", customErr.Error())
	}

	// 测试 Errors 字段
	if customErr.Errors() != nil {
		t.Errorf("期望 Errors 为 nil, 实际得到 %v", customErr.Errors())
	}
}

// 测试带详细错误的业务错误
func TestBizErrorWithDetails(t *testing.T) {
	errors := []interface{}{
		map[string]string{"field": "name", "message": "名称不能为空"},
		map[string]string{"field": "age", "message": "年龄必须大于0"},
	}
	customErr := NewBizErrorWithDetails("VALIDATION_ERROR", "验证失败", http.StatusBadRequest, errors)

	if customErr.Code() != "VALIDATION_ERROR" {
		t.Errorf("期望 code 为 'VALIDATION_ERROR', 实际得到 %v", customErr.Code())
	}

	if len(customErr.Errors()) != 2 {
		t.Errorf("期望 Errors 长度为 2, 实际得到 %d", len(customErr.Errors()))
	}
}

// 测试错误响应格式包含 errors 字段
func TestErrorResponseWithErrors(t *testing.T) {
	r := gin.New()

	handleFunc := func(ctx context.Context, req *testRequest) (*testResponse, error) {
		errors := []interface{}{
			map[string]string{"field": "id", "message": "ID格式错误"},
		}
		return nil, NewBizErrorWithDetails(40001, "参数验证失败", http.StatusBadRequest, errors)
	}

	r.GET("/test/:id", Handler(handleFunc))

	req := httptest.NewRequest("GET", "/test/123", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 %d, 实际得到 %d", http.StatusBadRequest, w.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	if resp.Message != "参数验证失败" {
		t.Errorf("期望 message 为 '参数验证失败', 实际得到 '%s'", resp.Message)
	}

	if len(resp.Errors) != 1 {
		t.Errorf("期望 Errors 长度为 1, 实际得到 %d", len(resp.Errors))
	}
}

// 测试 POST 请求 JSON body 绑定
func TestHandlerJSONBody(t *testing.T) {
	type createRequest struct {
		Name string `json:"name" binding:"required"`
		Age  int    `json:"age" binding:"required"`
	}

	type createResponse struct {
		ID      int64  `json:"id"`
		Name    string `json:"name"`
		Age     int    `json:"age"`
		Message string `json:"message"`
	}

	r := gin.New()

	handleFunc := func(ctx context.Context, req *createRequest) (*createResponse, error) {
		return &createResponse{
			ID:      12345,
			Name:    req.Name,
			Age:     req.Age,
			Message: "创建成功",
		}, nil
	}

	r.POST("/create", Handler(handleFunc))

	reqBody := map[string]interface{}{
		"name": "李四",
		"age":  30,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/create", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际得到 %d", http.StatusOK, w.Code)
	}

	var resp SuccessResponse[createResponse]
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	if resp.Data.Name != "李四" {
		t.Errorf("期望 Name 为 '李四', 实际得到 '%s'", resp.Data.Name)
	}

	if resp.Data.Age != 30 {
		t.Errorf("期望 Age 为 30, 实际得到 %d", resp.Data.Age)
	}
}

// 测试混合参数绑定（路径参数 + JSON body）
func TestHandlerMixedParams(t *testing.T) {
	type updateRequest struct {
		ID   int64  `path:"id"`
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	type updateResponse struct {
		ID      int64  `json:"id"`
		Name    string `json:"name"`
		Age     int    `json:"age"`
		Message string `json:"message"`
	}

	r := gin.New()

	handleFunc := func(ctx context.Context, req *updateRequest) (*updateResponse, error) {
		return &updateResponse{
			ID:      req.ID,
			Name:    req.Name,
			Age:     req.Age,
			Message: "更新成功",
		}, nil
	}

	r.PUT("/update/:id", Handler(handleFunc))

	reqBody := map[string]interface{}{
		"name": "王五",
		"age":  35,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/update/999", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际得到 %d", http.StatusOK, w.Code)
	}

	var resp SuccessResponse[updateResponse]
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	if resp.Data.ID != 999 {
		t.Errorf("期望 ID 为 999, 实际得到 %d", resp.Data.ID)
	}

	if resp.Data.Name != "王五" {
		t.Errorf("期望 Name 为 '王五', 实际得到 '%s'", resp.Data.Name)
	}

	if resp.Data.Age != 35 {
		t.Errorf("期望 Age 为 35, 实际得到 %d", resp.Data.Age)
	}
}

// 测试绑定验证错误返回详细错误信息
func TestBindValidationErrorWithDetails(t *testing.T) {
	type validationRequest struct {
		Name string `json:"name" binding:"required"`
		Age  int    `json:"age" binding:"required,min=1,max=150"`
	}

	type validationResponse struct {
		Message string `json:"message"`
	}

	r := gin.New()

	handleFunc := func(ctx context.Context, req *validationRequest) (*validationResponse, error) {
		return &validationResponse{Message: "success"}, nil
	}

	r.POST("/validate", Handler(handleFunc))

	// 发送一个缺少 required 字段的请求
	reqBody := map[string]interface{}{
		// 缺少 name 和 age 字段
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/validate", bytes.NewReader(bodyBytes))
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

	if resp.Message != "参数绑定失败" {
		t.Errorf("期望 message 为 '参数绑定失败', 实际得到 '%s'", resp.Message)
	}

	// 验证是否包含详细错误信息
	if resp.Errors == nil || len(resp.Errors) == 0 {
		t.Errorf("期望包含详细错误信息，但 errors 字段为空")
	}

	// 验证错误详情包含字段信息
	if len(resp.Errors) > 0 {
		// 至少应该有一个错误（可能是 Name 或 Age 字段）
		firstError, ok := resp.Errors[0].(map[string]interface{})
		if !ok {
			t.Errorf("期望错误详情为 map[string]interface{} 类型，实际得到 %T", resp.Errors[0])
		} else {
			if _, hasField := firstError["field"]; !hasField {
				t.Errorf("期望错误详情包含 'field' 字段")
			}
			if _, hasMessage := firstError["message"]; !hasMessage {
				t.Errorf("期望错误详情包含 'message' 字段")
			}
		}
	}
}
