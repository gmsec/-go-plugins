/*
	context 跟gin.context 的转换
*/
package api

import (
	"context"
)

// Context Wrapping gin context to custom context
type Context struct { // 包装gin的上下文到自定义context
	context.Context
}

// SetValue 设置key v
func (c *Context) SetValue(key interface{}, value interface{}) {
	ctx := context.WithValue(c.Context, key, value)
	c.Context = ctx
}

// GetValue 获取key
func (c *Context) GetValue(key interface{}) interface{} {
	return c.Context.Value(key)
}
