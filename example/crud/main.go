package main

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	handler "github.com/night1008/gotools/gin-api-handler"
)

// 示例请求结构
type GetUserRequest struct {
	UserID int64  `path:"id"`   // 从路径参数绑定
	Name   string `json:"name"` // 从 JSON body 绑定
	Age    int    `form:"age"`  // 从 query 参数绑定
}

// 示例响应结构
type GetUserResponse struct {
	UserID  int64  `json:"user_id"`
	Name    string `json:"name"`
	Age     int    `json:"age"`
	Message string `json:"message"`
}

// 业务处理函数
func handleGetUser(ctx context.Context, req *GetUserRequest) (*GetUserResponse, error) {
	// 模拟业务逻辑
	if req.UserID == 0 {
		return nil, handler.ErrBadRequest(40000, "用户ID不能为空")
	}

	if req.UserID == 999 {
		return nil, handler.ErrNotFound(40400, "用户不存在")
	}

	// 返回成功响应
	return &GetUserResponse{
		UserID:  req.UserID,
		Name:    req.Name,
		Age:     req.Age,
		Message: fmt.Sprintf("获取用户 %d 信息成功", req.UserID),
	}, nil
}

// 创建用户请求
type CreateUserRequest struct {
	Name string `json:"name" binding:"required"`
	Age  int    `json:"age" binding:"required,min=1,max=150"`
}

// 创建用户响应
type CreateUserResponse struct {
	UserID  int64  `json:"user_id"`
	Message string `json:"message"`
}

// 创建用户业务逻辑
func handleCreateUser(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error) {
	// 模拟参数验证
	var validationErrors []interface{}
	if req.Age < 18 {
		validationErrors = append(validationErrors, map[string]string{
			"field":   "age",
			"message": "年龄必须大于18岁",
		})
	}
	if len(req.Name) < 2 {
		validationErrors = append(validationErrors, map[string]string{
			"field":   "name",
			"message": "姓名长度必须大于2",
		})
	}

	if len(validationErrors) > 0 {
		return nil, handler.NewBizErrorWithDetails(
			"VALIDATION_ERROR",
			"参数验证失败",
			400,
			validationErrors,
		)
	}

	// 返回成功响应
	return &CreateUserResponse{
		UserID:  12345,
		Message: fmt.Sprintf("用户 %s 创建成功", req.Name),
	}, nil
}

// 更新用户请求
type UpdateUserRequest struct {
	UserID int64  `path:"id"`
	Name   string `json:"name"`
	Age    int    `json:"age"`
}

// 更新用户响应
type UpdateUserResponse struct {
	Message string `json:"message"`
}

// 更新用户业务逻辑
func handleUpdateUser(ctx context.Context, req *UpdateUserRequest) (*UpdateUserResponse, error) {
	if req.UserID == 0 {
		return nil, handler.ErrBadRequest(40000, "用户ID不能为空")
	}

	return &UpdateUserResponse{
		Message: fmt.Sprintf("用户 %d 更新成功", req.UserID),
	}, nil
}

// 删除用户请求
type DeleteUserRequest struct {
	UserID uint64 `path:"id"` // 测试 uint64 类型
}

// 删除用户响应
type DeleteUserResponse struct {
	Message string `json:"message"`
}

// 删除用户业务逻辑
func handleDeleteUser(ctx context.Context, req *DeleteUserRequest) (*DeleteUserResponse, error) {
	if req.UserID == 0 {
		return nil, handler.ErrBadRequest(40000, "用户ID不能为空")
	}

	return &DeleteUserResponse{
		Message: fmt.Sprintf("用户 %d 删除成功", req.UserID),
	}, nil
}

func main() {
	r := gin.Default()

	// 注册路由，使用通用 Handler
	r.GET("/user/:id", handler.Handler(handleGetUser))
	r.POST("/user", handler.Handler(handleCreateUser))
	r.PUT("/user/:id", handler.Handler(handleUpdateUser))
	r.DELETE("/user/:id", handler.Handler(handleDeleteUser))

	// 启动服务
	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}
