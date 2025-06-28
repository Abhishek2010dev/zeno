package zeno

import (
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
	params := make(map[string]string)
	for i, n := range c.pnames {
		if i < len(c.pvalues) {
			params[n] = c.pvalues[i]
		}
	}
	return params
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
