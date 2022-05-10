package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"unsafe"

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

	mylog.Info("using default gin.context")
	r, _ := gin.CreateTestContext(httptest.NewRecorder())
	r.Request = &http.Request{
		Method: "POST",
		Header: GetKeyValues(c),
	}
	return r
}

type iface struct{ itab, data uintptr }
type valueCtx struct {
	context.Context
	key, val interface{}
}

func GetKeyValues(ctx context.Context) map[string][]string {
	m := make(map[string][]string)
	getKeyValue(ctx, m)
	return m
}
func getKeyValue(ctx context.Context, m map[string][]string) {
	ictx := *(*iface)(unsafe.Pointer(&ctx))
	if ictx.data == 0 {
		return
	}
	valCtx := (*valueCtx)(unsafe.Pointer(ictx.data))
	if valCtx != nil && valCtx.key != nil && valCtx.val != nil {
		key, ok := valCtx.key.(string)
		if ok {
			val, ok := valCtx.val.(string)
			if ok {
				m[key] = append(m[key], val)
			} else {
				vals, ok := valCtx.val.([]string)
				if ok {
					m[key] = append(m[key], vals...)
				}
			}
		}

	}
	getKeyValue(valCtx.Context, m)
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
