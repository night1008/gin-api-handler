package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	handler "github.com/night1008/gotools/gin-api-handler"
)

// 用户注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Email    string `json:"email" binding:"required,email"`
	Age      int    `json:"age" binding:"required,min=18,max=100"`
}

// 用户注册响应
type RegisterResponse struct {
	UserID  int64  `json:"user_id"`
	Message string `json:"message"`
}

// 注册业务逻辑
func handleRegister(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	// 模拟业务逻辑
	return &RegisterResponse{
		UserID:  12345,
		Message: fmt.Sprintf("用户 %s 注册成功", req.Username),
	}, nil
}

func main() {
	r := gin.Default()

	// 方法 1: 使用默认配置（从 Accept-Language 头自动获取语言）
	r.POST("/register", handler.Handler(handleRegister))

	// 方法 2: 强制使用英文
	englishTranslator := handler.NewSimpleTranslator("en")
	r.POST("/register/en", handler.Handler(handleRegister,
		handler.WithTranslator(englishTranslator),
	))

	// 方法 3: 强制使用中文
	chineseTranslator := handler.NewSimpleTranslator("zh")
	r.POST("/register/zh", handler.Handler(handleRegister,
		handler.WithTranslator(chineseTranslator),
	))

	// 方法 4: 从 query 参数获取语言
	customLocaleFunc := func(r *http.Request) string {
		lang := r.URL.Query().Get("lang")
		if lang == "" {
			return "zh"
		}
		return lang
	}
	r.POST("/register/custom", handler.Handler(handleRegister,
		handler.WithLocaleFunc(customLocaleFunc),
	))

	fmt.Println("服务已启动在 :8080")
	fmt.Println("测试端点：")
	fmt.Println("  POST /register - 自动检测语言（从 Accept-Language 头）")
	fmt.Println("  POST /register/en - 强制英文")
	fmt.Println("  POST /register/zh - 强制中文")
	fmt.Println("  POST /register/custom?lang=en - 从 query 参数获取语言")
	fmt.Println()
	fmt.Println("测试命令示例：")
	fmt.Println("  # 中文错误（默认）")
	fmt.Println(`  curl -X POST http://localhost:8080/register -H "Content-Type: application/json" -d '{"username":"ab"}' | jq`)
	fmt.Println()
	fmt.Println("  # 英文错误（通过 Accept-Language 头）")
	fmt.Println(`  curl -X POST http://localhost:8080/register -H "Content-Type: application/json" -H "Accept-Language: en" -d '{"username":"ab"}' | jq`)
	fmt.Println()
	fmt.Println("  # 英文错误（通过路径）")
	fmt.Println(`  curl -X POST http://localhost:8080/register/en -H "Content-Type: application/json" -d '{"username":"ab"}' | jq`)
	fmt.Println()
	fmt.Println("  # 英文错误（通过 query 参数）")
	fmt.Println(`  curl -X POST "http://localhost:8080/register/custom?lang=en" -H "Content-Type: application/json" -d '{"username":"ab"}' | jq`)

	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}
