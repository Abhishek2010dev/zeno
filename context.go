package zeno

import (
	"mime/multipart"
	"net"

	"github.com/valyala/fasthttp"
)

type Context struct {
	RequestCtx *fasthttp.RequestCtx

	zeno     *Zeno
	pnames   []string
	pvalues  []string
	index    int
	handlers []Handler
}

func (c *Context) Next() error {
	c.index++
	for n := len(c.handlers); c.index < n; c.index++ {
		if err := c.handlers[c.index](c); err != nil {
			return err
		}
	}
	return nil
}

func (c *Context) Abort() {
	c.index = len(c.handlers)
}

func (c *Context) URL(route string, pairs ...interface{}) string {
	if r := c.zeno.routes[route]; r != nil {
		return r.URL(pairs...)
	}
	return ""
}

func (c *Context) init(ctx *fasthttp.RequestCtx) {
	c.RequestCtx = ctx
	c.index = -1
}

func (c *Context) Zeno() *Zeno {
	return c.zeno
}

func (c *Context) Status(code int) *Context {
	c.RequestCtx.SetStatusCode(code)
	return c
}

func (c *Context) SendString(value string) error {
	c.RequestCtx.Response.AppendBodyString(value)
	return nil
}

func (c *Context) Param(name string) string {
	for i, n := range c.pnames {
		if n == name {
			return c.pvalues[i]
		}
	}
	return ""
}

func (c *Context) Params() map[string]string {
	params := map[string]string{}
	for i, n := range c.pnames {
		if i < len(c.pvalues) {
			params[n] = c.pvalues[i]
		}
	}
	return params
}

func (c *Context) Query(key string) string {
	return c.zeno.toString(c.RequestCtx.QueryArgs().Peek(key))
}

func (c *Context) QueryArray(key string) []string {
	args := c.RequestCtx.QueryArgs().PeekMulti(key)
	arr := make([]string, len(args))
	for i, b := range args {
		arr[i] = c.zeno.toString(b)
	}
	return arr
}

func (c *Context) QueryMap() map[string]string {
	m := map[string]string{}
	c.RequestCtx.QueryArgs().VisitAll(func(key, value []byte) {
		m[c.zeno.toString(key)] = c.zeno.toString(value)
	})
	return m
}

func (c *Context) Method() string {
	return c.zeno.toString(c.RequestCtx.Method())
}

func (c *Context) Path() string {
	return c.zeno.toString(c.RequestCtx.Path())
}

func (c *Context) Port() string {
	_, port, err := net.SplitHostPort(c.RequestCtx.RemoteAddr().String())
	if err != nil {
		return ""
	}
	return port
}

func (c *Context) GetHeader(key string) string {
	return c.zeno.toString(c.RequestCtx.Request.Header.Peek(key))
}

func (c *Context) HeaderMap() map[string]string {
	m := map[string]string{}
	c.RequestCtx.Request.Header.VisitAll(func(key, value []byte) {
		m[c.zeno.toString(key)] = c.zeno.toString(value)
	})
	return m
}

func (c *Context) FormValue(key string) string {
	return c.zeno.toString(c.RequestCtx.FormValue(key))
}

func (c *Context) FormFile(key string) (*multipart.FileHeader, error) {
	return c.RequestCtx.FormFile(key)
}

func (c *Context) MultipartForm() (*multipart.Form, error) {
	return c.RequestCtx.MultipartForm()
}
