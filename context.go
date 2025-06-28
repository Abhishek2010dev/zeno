package zeno

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net"
	"sort"
	"strconv"
	"strings"

	"github.com/valyala/fasthttp"
)

// Context represents the context of the current HTTP request.
// It holds request and response data, route parameters, and provides
// convenience methods for handling various aspects of the request lifecycle.
type Context struct {
	// RequestCtx is the underlying fasthttp request context.
	RequestCtx *fasthttp.RequestCtx

	zeno     *Zeno
	pnames   []string
	pvalues  []string
	index    int
	handlers []Handler
}

// Next executes the next handler in the middleware chain.
// It returns early if any handler returns an error.
func (c *Context) Next() error {
	c.index++
	for n := len(c.handlers); c.index < n; c.index++ {
		if err := c.handlers[c.index](c); err != nil {
			return err
		}
	}
	return nil
}

// Abort stops the execution of any remaining middleware/handlers.
func (c *Context) Abort() {
	c.index = len(c.handlers)
}

// URL returns a URL for a named route with optional path parameters.
func (c *Context) URL(route string, pairs ...any) string {
	if r := c.zeno.routes[route]; r != nil {
		return r.URL(pairs...)
	}
	return ""
}

// init prepares the context with a new fasthttp.RequestCtx.
func (c *Context) init(ctx *fasthttp.RequestCtx) {
	c.RequestCtx = ctx
	c.index = -1
}

// Zeno returns the underlying Zeno engine instance.
func (c *Context) Zeno() *Zeno {
	return c.zeno
}

// Status sets the HTTP status code for the response.
func (c *Context) Status(code int) *Context {
	c.RequestCtx.SetStatusCode(code)
	return c
}

// SendString writes a plain text response body.
func (c *Context) SendString(value string) error {
	c.RequestCtx.Response.AppendBodyString(value)
	return nil
}

// Param returns the value of a route parameter by name.
//
// If the parameter is not present and a defaultValue is provided,
// the first element of defaultValue is returned instead.
//
// Example usage:
//
//	id := ctx.Param("id")              // returns "" if not found
//	id := ctx.Param("id", "default")   // returns "default" if not found
func (c *Context) Param(name string, defaultValue ...string) string {
	for i, n := range c.pnames {
		if n == name {
			return c.pvalues[i]
		}
	}
	if 0 < len(defaultValue) {
		return defaultValue[0]
	}
	return ""
}

// Params returns a map of all route parameters.
func (c *Context) Params() map[string]string {
	params := map[string]string{}
	for i, n := range c.pnames {
		if i < len(c.pvalues) {
			params[n] = c.pvalues[i]
		}
	}
	return params
}

// Query returns the query parameter value for the given key.
//
// If the parameter is not present and a defaultValue is provided,
// the first element of defaultValue is returned instead.
//
// Example usage:
//
//	name := ctx.Query("name")                   // returns "" if not found
//	name := ctx.Query("name", "default-name")   // returns "default-name" if not found
func (c *Context) Query(key string, defaultValue ...string) string {
	val := c.RequestCtx.QueryArgs().Peek(key)
	if len(val) == 0 && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return c.zeno.toString(val)
}

// QueryArray returns all query values for a given key.
func (c *Context) QueryArray(key string) []string {
	args := c.RequestCtx.QueryArgs().PeekMulti(key)
	arr := make([]string, len(args))
	for i, b := range args {
		arr[i] = c.zeno.toString(b)
	}
	return arr
}

// QueryMap returns all query parameters as a map.
func (c *Context) QueryMap() map[string]string {
	m := map[string]string{}
	c.RequestCtx.QueryArgs().VisitAll(func(key, value []byte) {
		m[c.zeno.toString(key)] = c.zeno.toString(value)
	})
	return m
}

// Method returns the HTTP method of the request.
func (c *Context) Method() string {
	return c.zeno.toString(c.RequestCtx.Method())
}

// Path returns the request URL path.
func (c *Context) Path() string {
	return c.zeno.toString(c.RequestCtx.Path())
}

// Port returns the remote port from the client's address.
func (c *Context) Port() string {
	_, port, err := net.SplitHostPort(c.RequestCtx.RemoteAddr().String())
	if err != nil {
		return ""
	}
	return port
}

// IP returns the remote IP address of the client.
func (c *Context) IP() string {
	return c.RequestCtx.RemoteIP().String()
}

// GetForwardedIPs returns a slice of IPs from the X-Forwarded-For header.
func (c *Context) GetForwardedIPs() []string {
	xForwardedFor := c.GetHeader(HeaderForwardedFor)
	if xForwardedFor == "" {
		return nil
	}
	parts := strings.Split(xForwardedFor, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

// GetHeader returns the value of the specified request header.
func (c *Context) GetHeader(key string) string {
	return c.zeno.toString(c.RequestCtx.Request.Header.Peek(key))
}

// RealIP returns the client's real IP address, considering X-Forwarded-For.
func (c *Context) RealIP() string {
	xForwardedFor := c.GetHeader(HeaderForwardedFor)
	if xForwardedFor != "" {
		parts := strings.Split(xForwardedFor, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	return c.IP()
}

// HeaderMap returns all request headers as a map.
func (c *Context) HeaderMap() map[string]string {
	m := map[string]string{}
	c.RequestCtx.Request.Header.VisitAll(func(key, value []byte) {
		m[c.zeno.toString(key)] = c.zeno.toString(value)
	})
	return m
}

// FormValue returns the value of a form field or a default if not present.
func (c *Context) FormValue(key string, defaultValue ...string) string {
	val := c.RequestCtx.FormValue(key)
	if len(val) == 0 && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return c.zeno.toString(val)
}

// FormFile returns the uploaded file header for the given form key.
func (c *Context) FormFile(key string) (*multipart.FileHeader, error) {
	return c.RequestCtx.FormFile(key)
}

// MultipartForm returns the parsed multipart form data.
func (c *Context) MultipartForm() (*multipart.Form, error) {
	return c.RequestCtx.MultipartForm()
}

// Body returns the raw request body.
func (c *Context) Body() []byte {
	return c.RequestCtx.Request.Body()
}

// PostBody returns the POST request body.
func (c *Context) PostBody() []byte {
	return c.RequestCtx.PostBody()
}

// IsAJAX returns true if the request was made via AJAX.
func (c *Context) IsAJAX() bool {
	return c.GetHeader("X-Requested-With") == "XMLHttpRequest"
}

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
			if acc.value == offer || acc.value == "*" {
				return offers[i]
			}
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

// Accepts returns the best match from the offers based on the Accept header.
func (c *Context) Accepts(offers ...string) string {
	return matchAccept(c.GetHeader(HeaderAccept), offers)
}

// AcceptsCharset returns the best match from the offers based on Accept-Charset.
func (c *Context) AcceptsCharset(offers ...string) string {
	return matchAccept(c.GetHeader(HeaderAcceptCharset), offers)
}

// AcceptsEncoding returns the best match from the offers based on Accept-Encoding.
func (c *Context) AcceptsEncoding(offers ...string) string {
	return matchAccept(c.GetHeader(HeaderAcceptEncoding), offers)
}

// AcceptsLanguage returns the best match from the offers based on Accept-Language.
func (c *Context) AcceptsLanguage(offers ...string) string {
	return matchAccept(c.GetHeader(HeaderAcceptLanguage), offers)
}

// Protocol returns the request protocol version (e.g., HTTP/1.1).
func (c *Context) Protocol() string {
	return c.zeno.toString(c.RequestCtx.Request.Header.Protocol())
}

// Scheme returns the request scheme, "http" or "https".
func (c *Context) Scheme() string {
	if c.RequestCtx.IsTLS() {
		return "https"
	}
	return "http"
}

// IsSecure returns true if the request is over HTTPS.
func (c *Context) IsSecure() bool {
	return c.RequestCtx.IsTLS()
}

// HTTPRange represents a parsed byte range from the Range header.
type HTTPRange struct {
	Start, End int64
}

// Range represents a collection of HTTP byte ranges with unit type.
type Range struct {
	Type   string
	Ranges []HTTPRange
}

// Ranges parses the Range header and returns validated byte ranges.
func (c *Context) Ranges(maxSize int64) (*Range, error) {
	header := c.GetHeader("Range")
	if header == "" {
		return nil, errors.New("no Range header")
	}

	parts := strings.SplitN(header, "=", 2)
	if len(parts) != 2 {
		return nil, errors.New("invalid Range header format")
	}

	unit := strings.TrimSpace(parts[0])
	if unit != "bytes" {
		return nil, fmt.Errorf("unsupported range unit: %s", unit)
	}

	rangesSpec := parts[1]
	rangesStrs := strings.Split(rangesSpec, ",")
	var ranges []HTTPRange

	for _, rStr := range rangesStrs {
		rStr = strings.TrimSpace(rStr)
		bounds := strings.SplitN(rStr, "-", 2)
		if len(bounds) != 2 {
			return nil, fmt.Errorf("invalid range segment: %s", rStr)
		}

		var start, end int64
		var err error

		if bounds[0] == "" {
			end, err = strconv.ParseInt(bounds[1], 10, 64)
			if err != nil || end <= 0 {
				return nil, fmt.Errorf("invalid suffix range value: %s", rStr)
			}
			if end > maxSize {
				end = maxSize
			}
			start = max(maxSize-end, 0)
			end = maxSize - 1
		} else {
			start, err = strconv.ParseInt(bounds[0], 10, 64)
			if err != nil || start < 0 {
				return nil, fmt.Errorf("invalid start range value: %s", rStr)
			}
			if bounds[1] != "" {
				end, err = strconv.ParseInt(bounds[1], 10, 64)
				if err != nil || end < start {
					return nil, fmt.Errorf("invalid end range value: %s", rStr)
				}
				if end >= maxSize {
					end = maxSize - 1
				}
			} else {
				end = maxSize - 1
			}
		}

		if start >= maxSize || start > end {
			continue
		}

		ranges = append(ranges, HTTPRange{Start: start, End: end})
	}

	if len(ranges) == 0 {
		return nil, errors.New("no valid byte ranges found in header")
	}

	return &Range{
		Type:   unit,
		Ranges: ranges,
	}, nil
}

// max returns the maximum of two int64 values.
func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
