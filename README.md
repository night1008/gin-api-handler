# gin-api-handler

一个通用的 Gin API 处理器包，简化 API 开发，让开发者专注于业务逻辑。

## 特性

- **通用处理函数**：使用泛型定义统一的处理函数接口
- **自动参数绑定**：支持从路径、JSON、Query 参数自动绑定
- **路径参数支持**：通过 `path` tag 支持 int64、uint64、string 类型的路径参数
- **业务错误机制**：统一的业务错误处理和 HTTP 状态码映射
- **类型安全**：基于 Go 泛型，提供完整的类型安全
- **灵活配置**：支持函数式选项模式，可灵活指定成功码、HTTP 状态码等参数

## 要求

- Go 1.25.6+

## 安装

```bash
go get github.com/night1008/gotools/gin-api-handler
```

## 快速开始

### 1. 定义请求和响应结构

```go
type GetUserRequest struct {
    UserID int64  `path:"id"`     // 从路径参数绑定
    Name   string `json:"name"`   // 从 JSON body 绑定
    Age    int    `form:"age"`    // 从 query 参数绑定
}

type GetUserResponse struct {
    UserID  int64  `json:"user_id"`
    Name    string `json:"name"`
    Age     int    `json:"age"`
    Message string `json:"message"`
}
```

### 2. 实现业务逻辑

```go
func handleGetUser(ctx context.Context, req *GetUserRequest) (*GetUserResponse, error) {
    if req.UserID == 0 {
        return nil, handler.ErrBadRequest(40000, "用户ID不能为空")
    }

    if req.UserID == 999 {
        return nil, handler.ErrNotFound(40400, "用户不存在")
    }

    return &GetUserResponse{
        UserID:  req.UserID,
        Name:    req.Name,
        Age:     req.Age,
        Message: fmt.Sprintf("获取用户 %d 信息成功", req.UserID),
    }, nil
}
```

### 3. 注册路由

```go
r := gin.Default()

// 使用默认配置
r.GET("/user/:id", handler.Handler(handleGetUser))

// 使用函数式选项
r.POST("/user", handler.Handler(handleCreateUser,
    handler.WithSuccessCode(1),
    handler.WithSuccessHTTPCode(http.StatusCreated),
))
```

## 使用方式

### 1. 使用默认配置

```go
// 默认: code=0, httpCode=200, bindErrorCode=400
r.GET("/user/:id", handler.Handler(handleGetUser))
```

### 2. 使用函数式选项（推荐）

```go
// 只指定成功码
r.GET("/user/:id", handler.Handler(handleGetUser,
    handler.WithSuccessCode(1),
))

// 只指定 HTTP 状态码
r.POST("/user", handler.Handler(handleCreateUser,
    handler.WithSuccessHTTPCode(http.StatusCreated),
))

// 指定多个参数（任意组合）
r.POST("/article", handler.Handler(handleCreateArticle,
    handler.WithSuccessCode("OK"),
    handler.WithSuccessHTTPCode(http.StatusCreated),
    handler.WithBindErrorCode(40000),
))
```

### 3. 使用配置对象

```go
config := &handler.HandlerConfig{
    SuccessCode:     0,
    SuccessHTTPCode: http.StatusOK,
    BindErrorCode:   40000,
}
r.GET("/user/:id", handler.HandlerWithConfig(handleGetUser, config))
```

### 4. 直接指定所有参数

```go
r.GET("/user/:id", handler.HandlerWithCode(
    handleGetUser,
    0,                        // successCode
    http.StatusOK,           // successHTTPCode
    40000,                   // bindErrorCode
))
```

### 5. 修改全局默认配置

```go
// 修改全局默认配置
handler.DefaultConfig.SuccessCode = 1
handler.DefaultConfig.SuccessHTTPCode = http.StatusOK
handler.DefaultConfig.BindErrorCode = 40000

// 之后所有使用 Handler() 的路由都会使用新的默认配置
r.GET("/user/:id", handler.Handler(handleGetUser))
```

## 支持的参数绑定

### 路径参数（path tag）

通过 `path` tag 从 URL 路径中绑定参数，支持以下类型：

- `int64`
- `uint64`
- `string`

```go
type Request struct {
    UserID int64  `path:"id"`      // /user/123
    Type   string `path:"type"`    // /user/123/profile
}
```

### 其他参数

使用 Gin 的标准 tag：

- `json` - 从 JSON body 绑定
- `form` - 从 query 参数或 form 绑定
- `uri` - 从 URI 绑定
- `header` - 从 HTTP header 绑定

## 业务错误处理

### 错误响应格式

业务错误会自动转换为统一的 JSON 响应：

```json
{
    "code": 40400,
    "message": "用户不存在"
}
```

带详细错误信息的响应：

```json
{
    "code": "VALIDATION_ERROR",
    "message": "参数验证失败",
    "errors": [
        {"field": "name", "message": "姓名不能为空"},
        {"field": "age", "message": "年龄必须大于0"}
    ]
}
```

### 使用预定义错误

所有预定义错误都支持自定义错误码和消息：

```go
return nil, handler.ErrBadRequest(40000, "参数错误")
return nil, handler.ErrUnauthorized(40100, "未授权")
return nil, handler.ErrForbidden(40300, "禁止访问")
return nil, handler.ErrNotFound(40400, "资源不存在")
return nil, handler.ErrConflict(40900, "资源冲突")
return nil, handler.ErrInternalServer(50000, "内部错误")
```

预定义错误与 HTTP 状态码的映射：
- `ErrBadRequest` → 400
- `ErrUnauthorized` → 401
- `ErrForbidden` → 403
- `ErrNotFound` → 404
- `ErrConflict` → 409
- `ErrInternalServer` → 500

### 自定义业务错误

```go
// 简单错误（code 支持 any 类型）
customErr := handler.NewBizError(
    10001,                    // 业务错误码（可以是 int、string 等）
    "自定义错误消息",           // 错误消息
    http.StatusBadRequest,    // HTTP 状态码
)
return nil, customErr

// 带详细错误信息
errors := []any{
    map[string]string{"field": "email", "message": "邮箱格式不正确"},
    map[string]string{"field": "phone", "message": "手机号格式不正确"},
}
customErr := handler.NewBizErrorWithDetails(
    "VALIDATION_ERROR",       // 错误码（支持字符串）
    "参数验证失败",             // 错误消息
    http.StatusBadRequest,    // HTTP 状态码
    errors,                   // 详细错误列表
)
return nil, customErr
```

### 错误响应格式

简单错误响应：

```json
{
    "code": 40400,
    "message": "用户不存在"
}
```

带详细错误的响应：

```json
{
    "code": "VALIDATION_ERROR",
    "message": "参数验证失败",
    "errors": [
        {"field": "email", "message": "邮箱格式不正确"},
        {"field": "phone", "message": "手机号格式不正确"}
    ]
}
```

## 成功响应格式

成功的响应会自动包装为统一格式：

```json
{
    "code": 0,
    "data": {
        "user_id": 123,
        "name": "张三",
        "age": 25,
        "message": "获取用户 123 信息成功"
    }
}
```

## 完整示例

```go
package main

import (
    "context"
    "github.com/gin-gonic/gin"
    handler "github.com/night1008/gotools/gin-api-handler"
)

type CreateUserRequest struct {
    Name string `json:"name" binding:"required"`
    Age  int    `json:"age" binding:"required,min=1,max=150"`
}

type CreateUserResponse struct {
    UserID  int64  `json:"user_id"`
    Message string `json:"message"`
}

func handleCreateUser(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error) {
    if req.Age < 18 {
        return nil, handler.NewBizError(40001, "用户年龄必须大于18岁", 400)
    }

    return &CreateUserResponse{
        UserID:  12345,
        Message: "用户创建成功",
    }, nil
}

func main() {
    r := gin.Default()
    r.POST("/user", handler.Handler(handleCreateUser))
    r.Run(":8080")
}
```

## 测试

运行测试：

```bash
go test -v
```

## API 文档

### 配置选项

#### Option

```go
type Option func(*HandlerConfig)
```

处理器选项函数类型。

#### WithSuccessCode

```go
func WithSuccessCode(code any) Option
```

设置成功响应的业务代码。

#### WithSuccessHTTPCode

```go
func WithSuccessHTTPCode(code int) Option
```

设置成功响应的 HTTP 状态码。

#### WithBindErrorCode

```go
func WithBindErrorCode(code any) Option
```

设置参数绑定错误的业务代码。

### 处理器函数

#### Handler

```go
func Handler[T any, R any](handleFunc HandleFunc[T, R], opts ...Option) gin.HandlerFunc
```

创建一个 Gin 处理器函数，支持函数式选项。

**参数：**
- `handleFunc` - 业务处理函数
- `opts` - 可选的配置选项

**返回：**
- `gin.HandlerFunc` - Gin 路由处理器

#### HandlerWithConfig

```go
func HandlerWithConfig[T any, R any](handleFunc HandleFunc[T, R], config *HandlerConfig) gin.HandlerFunc
```

使用配置对象创建处理器。

**参数：**
- `handleFunc` - 业务处理函数
- `config` - 处理器配置对象

#### HandlerWithCode

```go
func HandlerWithCode[T any, R any](handleFunc HandleFunc[T, R], successCode any, successHTTPCode int, bindErrorCode any) gin.HandlerFunc
```

直接指定所有参数创建处理器。

**参数：**
- `handleFunc` - 业务处理函数
- `successCode` - 成功响应的业务代码
- `successHTTPCode` - 成功响应的 HTTP 状态码
- `bindErrorCode` - 参数绑定错误的业务代码

### 类型定义

#### HandleFunc

```go
type HandleFunc[T any, R any] func(ctx context.Context, req *T) (*R, error)
```

通用业务处理函数类型。

**参数：**
- `ctx` - 上下文对象
- `req` - 请求对象指针

**返回：**
- `*R` - 响应对象指针
- `error` - 错误（可以是业务错误或普通错误）

#### BizError

```go
type BizError interface {
    error
    Code() any
    HTTPCode() int
    Errors() []any
}
```

业务错误接口。

#### HandlerConfig

```go
type HandlerConfig struct {
    SuccessCode     any
    SuccessHTTPCode int
    BindErrorCode   any
}
```

处理器配置结构。

### 错误函数

#### NewBizError

```go
func NewBizError(code any, message string, httpCode int) BizError
```

创建简单业务错误。

**参数：**
- `code` - 业务错误码（支持 any 类型）
- `message` - 错误消息
- `httpCode` - HTTP 状态码

#### NewBizErrorWithDetails

```go
func NewBizErrorWithDetails(code any, message string, httpCode int, errors []any) BizError
```

创建带详细错误的业务错误。

**参数：**
- `code` - 业务错误码
- `message` - 错误消息
- `httpCode` - HTTP 状态码
- `errors` - 详细错误列表

#### 预定义错误函数

```go
func ErrBadRequest(code any, msg string) BizError      // 400
func ErrUnauthorized(code any, msg string) BizError    // 401
func ErrForbidden(code any, msg string) BizError       // 403
func ErrNotFound(code any, msg string) BizError        // 404
func ErrConflict(code any, msg string) BizError        // 409
func ErrInternalServer(code any, msg string) BizError  // 500
```

## License

MIT
