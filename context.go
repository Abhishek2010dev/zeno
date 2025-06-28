package zeno

import (
	"mime/multipart"
	"net"
	"sort"
	"strconv"
	"strings"

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

func (c *Context) IP() string {
	return c.RequestCtx.RemoteIP().String()
}

func (c *Context) GetHeader(key string) string {
	return c.zeno.toString(c.RequestCtx.Request.Header.Peek(key))
}

func (c *Context) RealIP() string {
	ip := c.GetHeader("X-Real-IP")
	if len(ip) == 0 {
		ip = c.GetHeader("X-Forwarded-For")
	}
	if len(ip) == 0 {
		return c.IP()
	}
	return ip
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

func (c *Context) Body() []byte {
	return c.RequestCtx.Request.Body()
}

func (c *Context) PostBody() []byte {
	return c.RequestCtx.PostBody()
}

func (c *Context) IsAJAX() bool {
	return c.GetHeader("X-Requested-With") == "XMLHttpRequest"
}

// acceptItem is a helper struct for parsing Accept header values, including their quality factor (q).
type acceptItem struct {
	value string
	q     float64 // Quality factor
}

func parseAccept(header string) []acceptItem {
	parts := strings.Split(header, ",")
	items := make([]acceptItem, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		q := 1.0
		if idx := strings.Index(part, ";q="); idx != -1 {
			qValStr := part[idx+3:]
			part = part[:idx]
			if qVal, err := strconv.ParseFloat(qValStr, 64); err == nil {
				q = qVal
			}
		}
		items = append(items, acceptItem{value: strings.ToLower(part), q: q})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].q > items[j].q
	})

	return items
}

func matchAccept(header string, offers []string) string {
	if header == "" || len(offers) == 0 {
		return ""
	}

	accepted := parseAccept(header)
	offersLower := make([]string, len(offers))
	for i, o := range offers {
		offersLower[i] = strings.ToLower(o)
	}

	for _, acc := range accepted {
		for i, offer := range offersLower {
			// Exact match or wildcard *
			if acc.value == offer || acc.value == "*" {
				return offers[i]
			}
			// Type wildcard match (e.g., "text/*" matching "text/html")
			if strings.HasSuffix(acc.value, "/*") {
				prefix := strings.TrimSuffix(acc.value, "*")
				if strings.HasPrefix(offer, prefix) {
					return offers[i]
				}
			}
		}
	}
	return ""
}

// Accepts determines the best content type that the client accepts based on the
// Accept request header and the provided offers.
func (c *Context) Accepts(offers ...string) string {
	return matchAccept(c.GetHeader(HeaderAccept), offers)
}

// AcceptsCharset determines the best character set that the client accepts based on the
// Accept-Charset request header and the provided offers.
func (c *Context) AcceptsCharset(offers ...string) string {
	return matchAccept(c.Header(HeaderAcceptCharset), offers)
}

// AcceptsEncoding determines the best encoding that the client accepts based on the
// Accept-Encoding request header and the provided offers.
func (c *Context) AcceptsEncoding(offers ...string) string {
	return matchAccept(c.Header(HeaderAcceptEncoding), offers)
}

// AcceptsLanguage determines the best language that the client accepts based on the
// Accept-Language request header and the provided offers.
func (c *Context) AcceptsLanguage(offers ...string) string {
	return matchAccept(c.Header(HeaderAcceptLanguage), offers)
}
