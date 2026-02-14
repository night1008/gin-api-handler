## gin-api-handler

实现一个通用包，让用户不用重复写 handle 代码，主要实现以下

1. 通用接口

```go
type HandleFunc[T any, R any] func(ctx context.Context, req *T) (*R, error)
```

通过反射 req 中的 path tag，自动绑定路径中的参数，支持 int64, uint64, string 三种类型


2. 通用业务错误机制

允许用户自定义业务错误类型，可以在上面的 HandleFunc 中返回业务错误，统一在 handler 层处理，并返回给前端，结构为
```go
type Error struct {
	Code    interface{}   `json:"code"` // 为了适配 openapi 的 code 字段，这里使用 interface{}
	Message string        `json:"message"`
	Errors  []interface{} `json:"errors,omitempty"`
	Status  int           `json:"-"`
}
```
根据不同错误返回不同的 http code 和错误信息