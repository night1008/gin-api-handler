// Package main provides a Swagger example for gin-api-handler.
//
//	@title			Gin API Handler Example
//	@version		1.0
//	@description	A sample CRUD API using gin-api-handler with Swagger documentation.
//	@host			localhost:8080
//	@BasePath		/
package main

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	handler "github.com/night1008/gotools/gin-api-handler"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "example/swagger/docs"
)

// GetUserRequest represents the request parameters for getting a user.
type GetUserRequest struct {
	UserID int64  `path:"id"`
	Name   string `json:"name"`
	Age    int    `form:"age"`
}

// GetUserResponse represents the response for getting a user.
type GetUserResponse struct {
	UserID  int64  `json:"user_id"`
	Name    string `json:"name"`
	Age     int    `json:"age"`
	Message string `json:"message"`
}

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

// CreateUserRequest represents the request body for creating a user.
type CreateUserRequest struct {
	Name string `json:"name" binding:"required"`
	Age  int    `json:"age" binding:"required,min=1,max=150"`
}

// CreateUserResponse represents the response for creating a user.
type CreateUserResponse struct {
	UserID  int64  `json:"user_id"`
	Message string `json:"message"`
}

func handleCreateUser(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error) {
	return &CreateUserResponse{
		UserID:  12345,
		Message: fmt.Sprintf("用户 %s 创建成功", req.Name),
	}, nil
}

// UpdateUserRequest represents the request parameters for updating a user.
type UpdateUserRequest struct {
	UserID int64  `path:"id"`
	Name   string `json:"name"`
	Age    int    `json:"age"`
}

// UpdateUserResponse represents the response for updating a user.
type UpdateUserResponse struct {
	Message string `json:"message"`
}

func handleUpdateUser(ctx context.Context, req *UpdateUserRequest) (*UpdateUserResponse, error) {
	if req.UserID == 0 {
		return nil, handler.ErrBadRequest(40000, "用户ID不能为空")
	}

	return &UpdateUserResponse{
		Message: fmt.Sprintf("用户 %d 更新成功", req.UserID),
	}, nil
}

// DeleteUserRequest represents the request parameters for deleting a user.
type DeleteUserRequest struct {
	UserID uint64 `path:"id"`
}

// DeleteUserResponse represents the response for deleting a user.
type DeleteUserResponse struct {
	Message string `json:"message"`
}

func handleDeleteUser(ctx context.Context, req *DeleteUserRequest) (*DeleteUserResponse, error) {
	if req.UserID == 0 {
		return nil, handler.ErrBadRequest(40000, "用户ID不能为空")
	}

	return &DeleteUserResponse{
		Message: fmt.Sprintf("用户 %d 删除成功", req.UserID),
	}, nil
}

// GetUser godoc
//
//	@Summary		Get a user
//	@Description	Get user information by user ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int				true	"User ID"
//	@Success		200	{object}	GetUserResponse
//	@Failure		400	{object}	map[string]any
//	@Failure		404	{object}	map[string]any
//	@Router			/user/{id} [get]
func GetUser(c *gin.Context) {
	handler.Handler(handleGetUser)(c)
}

// CreateUser godoc
//
//	@Summary		Create a user
//	@Description	Create a new user with name and age
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			request	body		CreateUserRequest	true	"Create user request"
//	@Success		200		{object}	CreateUserResponse
//	@Failure		400		{object}	map[string]any
//	@Router			/user [post]
func CreateUser(c *gin.Context) {
	handler.Handler(handleCreateUser)(c)
}

// UpdateUser godoc
//
//	@Summary		Update a user
//	@Description	Update user information by user ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int					true	"User ID"
//	@Param			request	body		UpdateUserRequest	true	"Update user request"
//	@Success		200		{object}	UpdateUserResponse
//	@Failure		400		{object}	map[string]any
//	@Router			/user/{id} [put]
func UpdateUser(c *gin.Context) {
	handler.Handler(handleUpdateUser)(c)
}

// DeleteUser godoc
//
//	@Summary		Delete a user
//	@Description	Delete a user by user ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	DeleteUserResponse
//	@Failure		400	{object}	map[string]any
//	@Router			/user/{id} [delete]
func DeleteUser(c *gin.Context) {
	handler.Handler(handleDeleteUser)(c)
}

func main() {
	r := gin.Default()

	r.GET("/user/:id", GetUser)
	r.POST("/user", CreateUser)
	r.PUT("/user/:id", UpdateUser)
	r.DELETE("/user/:id", DeleteUser)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}
