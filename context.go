package zeno

import "github.com/valyala/fasthttp"

type Context struct {
	*fasthttp.RequestCtx
}
