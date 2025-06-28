package zeno

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func newTestApp() *Zeno {
	app := New()
	return app
}

func h(name string) Handler {
	return func(c *Context) error {
		return c.SendString(name)
	}
}

func TestRouteMatch_Static(t *testing.T) {
	app := newTestApp()

	app.Get("/hello", h("world"))

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/hello")
	ctx.Request.Header.SetMethod("GET")

	app.HandleRequest(ctx)
	assert.Equal(t, "world", string(ctx.Response.Body()))
	assert.Equal(t, 200, ctx.Response.StatusCode())
}

func TestRouteMatch_Params(t *testing.T) {
	app := newTestApp()

	app.Get("/users/{id}", func(c *Context) error {
		return c.SendString("User ID: " + c.Param("id"))
	})

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/users/42")
	ctx.Request.Header.SetMethod("GET")

	app.HandleRequest(ctx)
	assert.Equal(t, "User ID: 42", string(ctx.Response.Body()))
}

func TestRouteMatch_Regex(t *testing.T) {
	app := newTestApp()

	app.Get("/images/{file:[a-z]+\\.png}", func(c *Context) error {
		return c.SendString("File: " + c.Param("file"))
	})

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/images/logo.png")
	ctx.Request.Header.SetMethod("GET")

	app.HandleRequest(ctx)
	assert.Equal(t, "File: logo.png", string(ctx.Response.Body()))
}

func TestRouteMatch_Wildcard(t *testing.T) {
	app := newTestApp()

	app.Get("/static/*", func(c *Context) error {
		return c.SendString("Path: " + c.Param(":"))
	})

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/static/js/app.js")
	ctx.Request.Header.SetMethod("GET")

	app.HandleRequest(ctx)
	assert.Equal(t, "Path: js/app.js", string(ctx.Response.Body()))
}

func TestMethodNotAllowed(t *testing.T) {
	app := newTestApp()

	app.Get("/demo", h("ok"))

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/demo")
	ctx.Request.Header.SetMethod("POST")

	app.HandleRequest(ctx)
	assert.Equal(t, 405, ctx.Response.StatusCode())
	assert.Contains(t, string(ctx.Response.Header.Peek("Allow")), "GET")
}

func TestNotFound(t *testing.T) {
	app := newTestApp()

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/nope")
	ctx.Request.Header.SetMethod("GET")

	app.HandleRequest(ctx)
	assert.Equal(t, 404, ctx.Response.StatusCode())
}

func TestNamedRoute(t *testing.T) {
	app := newTestApp()

	route := app.Get("/users/{id}", h("ok")).Name("user.show")
	assert.Equal(t, route, app.routes["user.show"])

	url := route.URL("id", 123)
	assert.Equal(t, "/users/123", url)
}
