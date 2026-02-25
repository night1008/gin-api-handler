# Swagger Example

This example demonstrates how to integrate [Swagger](https://swagger.io/) documentation with `gin-api-handler`.

## Pattern

Swagger annotations are written on thin **gin handler wrapper functions**, not on business logic functions:

```go
// Business logic — no swagger annotations
func handleGetUser(ctx context.Context, req *GetUserRequest) (*GetUserResponse, error) {
    // ...
}

// GetUser godoc
//
//	@Summary		Get a user
//	@Tags			users
//	@Param			id	path		int				true	"User ID"
//	@Success		200	{object}	GetUserResponse
//	@Router			/user/{id} [get]
func GetUser(c *gin.Context) {
    handler.Handler(handleGetUser)(c)
}
```

## Setup

Install dependencies:

```bash
go mod tidy
```

## Generate Swagger Docs

Install the `swag` CLI tool:

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

Generate the docs (run from this directory):

```bash
swag init
```

This creates a `docs/` directory with `docs.go`, `swagger.json`, and `swagger.yaml`.

## Run

```bash
go run .
```

Then open <http://localhost:8080/swagger/index.html> in your browser.
