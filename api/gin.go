package api

import (
	"context"
	"net/http"
	"net/http/httptest"

	"github.com/xxjwxc/public/mylog"
	"github.com/xxjwxc/public/tools"
	"google.golang.org/grpc/metadata"

	"github.com/gin-gonic/gin"
)

// gin key
type ginHTTPReq struct{}

// GetVersion Get the version by req url
func (c *Context) GetVersion() string { // 获取版本号
	return c.GetGinCtx().Param("version")
}

//WriteJSON 写入json对象
func (c *Context) WriteJSON(obj interface{}) {
	c.GetGinCtx().JSON(200, obj)
}

// WriteHeadToCtx 设置头数据到grpc headre里面
func (c *Context) WriteHeadToCtx(g *gin.Context) {
	js := tools.JSONDecode(g.Request.Header)
	headerData := metadata.Pairs("gmsec-httpheader", js)
	c.Context = metadata.NewOutgoingContext(c.Context, headerData)
}

// AddHeadToCtx 添加一层数据到ctx里面
func (c *Context) AddHeadToCtx(kv ...string) { // 后续也可以往后面添加数据
	c.Context = metadata.AppendToOutgoingContext(c.Context, kv...)
}

// GetGinCtx 获取 gin.Context
func (c *Context) GetGinCtx() *gin.Context {
	req := c.GetValue(ginHTTPReq{})
	if req != nil {
		if r, ok := req.(*gin.Context); ok {
			return r
		}
	}

	mylog.Info("using default gin.context")
	r, _ := gin.CreateTestContext(httptest.NewRecorder())
	r.Request = &http.Request{
		Method: "POST",
		Header: GetKeyValues(c),
	}
	c.SetValue(ginHTTPReq{}, r)
	return r
}

func GetKeyValues(ctx context.Context) map[string][]string {
	m := make(map[string][]string)
	// 调用服务端方法的时候可以在后面传参数
	md, _ := metadata.FromIncomingContext(ctx)
	if v, ok := md["gmsec-httpheader"]; ok {
		if len(v) > 0 {
			tools.JSONEncode(v[0], &m)
		}
	}
	return m
}

// NewCtx Create a new custom context
func NewCtx(c *gin.Context) *Context { // 新建一个自定义context
	ctx := &Context{context.TODO()}
	ctx.SetValue(ginHTTPReq{}, c)
	return ctx
}

// NewAPIFunc default of custom handlefunc
func NewAPIFunc(c *gin.Context) interface{} {
	return NewCtx(c)
}
