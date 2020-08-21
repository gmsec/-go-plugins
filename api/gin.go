package api

import (
	"context"
	"net/http/httptest"

	"github.com/xxjwxc/public/mylog"

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

// GetGinCtx 获取 gin.Context
func (c *Context) GetGinCtx() *gin.Context {
	req := c.GetValue(ginHTTPReq{})
	if req != nil {
		if r, ok := req.(*gin.Context); ok {
			return r
		}
	}

	mylog.ErrorString("using default gin.context")
	r, _ := gin.CreateTestContext(httptest.NewRecorder())
	return r
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
