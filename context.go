package zeno

import (
	"bytes"
	"errors"
	"fmt"
	"mime/multipart"
	"net"
	"sort"
	"strconv"
	"strings"

	"github.com/bytedance/sonic"
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
	c.RequestCtx.Response.SetBodyString(value)
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

// SetHeader sets the HTTP response header with the given key and value.
func (c *Context) SetHeader(key, value string) {
	c.RequestCtx.Response.Header.Set(key, value)
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

// SendBytes sets the response body to the given byte slice `b`.
// It overwrites any previously set body content.
//
// This method is typically used when you already have the response
// body as a raw byte slice, such as when serving JSON, Yaml, or binary data.
//
// Example:
//
//	err := ctx.SendBytes([]byte("Hello, World!"))
//	if err != nil {
//	    // handle error
//	}
func (c *Context) SendBytes(b []byte) error {
	c.RequestCtx.Response.SetBodyRaw(b)
	return nil
}

// SendStatusCode sets the HTTP response status code to the given `code`.
// It does not modify the response body unless the body is currently empty,
// in which case it sets the body to the default status text.
//
// Example:
//
//	ctx.SendStatusCode(fasthttp.StatusNoContent) // sets 204 No Content
func (c *Context) SendStatusCode(code int) error {
	c.RequestCtx.SetStatusCode(code)
	if len(c.RequestCtx.Response.Body()) == 0 {
		return c.SendString(StatusMessage(code))
	}
	return nil
}

// SetContentType sets the “Content‑Type” response header.
//
// The supplied value must be a valid MIME type string, e.g. "application/json"
// or "text/html; charset=utf-8".  It is typically called by higher‑level
// helpers so callers do not need to set the header manually.
func (c *Context) SetContentType(value string) {
	c.SetHeader(HeaderContentType, value)
}

// SendHTML writes an HTML payload to the client.
//
// The helper first sets the Content‑Type header to
// "text/html; charset=utf-8", then delegates to SendString to transmit the
// body.  Useful for quick inline responses or template output.
//
// Example:
//
//	if err := ctx.SendHTML("<h1>Hello, Zeno!</h1>"); err != nil {
//	    // handle error
//	}
func (c *Context) SendHTML(value string) error {
	c.SetContentType("text/html; charset=utf-8")
	return c.SendString(value)
}

// SendFile streams the file located at path to the client.
//
// It uses fasthttp’s zero‑copy RequestCtx.SendFile under the hood, allowing the
// kernel to send the file directly from disk to the socket for maximal
// performance.  The Content‑Type header is inferred from the file extension
// via fasthttp’s built‑in MIME database.
//
// SendFile always returns nil because any I/O errors are already handled by
// fasthttp and reflected in the HTTP response it writes.  If you need to
// inspect those errors programmatically, call ctx.RequestCtx.SendFile
// directly and check the returned error.
//
// Example:
//
//	// Deliver a static asset.
//	_ = ctx.SendFile("./public/index.html")
func (c *Context) SendFile(path string) error {
	c.RequestCtx.SendFile(path)
	return nil
}

// SendHeader sets a response header with the given key and value.
// It returns nil for compatibility with middleware chains.
//
// Example:
//
//	err := ctx.SendHeader("X-Custom-Header", "value")
func (c *Context) SendHeader(key, value string) error {
	c.SetHeader(key, value)
	return nil
}

// SendJSON encodes the given value as JSON and writes it to the response.
// It sets the Content-Type to "application/json" unless overridden with the optional ctype argument.
//
// Example:
//
//	return c.SendJSON(data)
//	return c.SendJSON(data, "application/vnd.api+json")
func (c *Context) SendJSON(value any, ctype ...string) error {
	contentType := "application/json"
	if len(ctype) > 0 {
		contentType = ctype[0]
	}
	c.SetContentType(contentType)

	bytes, err := c.Zeno().JsonEncoder(value)
	if err != nil {
		return NewHTTPError(StatusInternalServerError, "Failed to encode JSON: "+err.Error())
	}
	return c.SendBytes(bytes)
}

// BindJSON decodes the JSON request body into the provided destination structure.
// Returns an error if the body is empty or invalid.
//
// Example:
//
//	var req UserInput
//	if err := c.BindJSON(&req); err != nil {
//	    return err
//	}
func (c *Context) BindJSON(out any) error {
	body := c.PostBody()
	if len(body) == 0 {
		return NewHTTPError(StatusBadRequest, "Request body is empty")
	}
	if err := c.Zeno().JsonDecoder(body, out); err != nil {
		return NewHTTPError(StatusBadRequest, "Invalid JSON: "+err.Error())
	}
	return nil
}

// SendJSONP encodes the value as JSON and wraps it in a JavaScript function call
// for use with JSONP (JSON with Padding). It sets Content-Type to "application/javascript".
//
// The optional callback name defaults to "callback".
//
// Example:
//
//	return c.SendJSONP(data)               // callback(data)
//	return c.SendJSONP(data, "handleData") // handleData(data)
func (c *Context) SendJSONP(value any, callback ...string) error {
	cback := "callback"
	if len(callback) > 0 {
		cback = callback[0]
	}
	c.SetContentType("application/javascript")
	bytes, err := c.Zeno().JsonEncoder(value)
	if err != nil {
		return NewHTTPError(StatusInternalServerError, "Failed to encode JSON: "+err.Error())
	}
	// Wrap the JSON in the callback function
	return c.SendString(cback + "(" + c.zeno.toString(bytes) + ");")
}

// SendPrettyJSON encodes the given value as pretty-formatted JSON (indented)
// and writes it to the response. Ideal for human-readable responses in development.
//
// It sets the Content-Type to "application/json" unless overridden.
//
// Example:
//
//	return c.SendPrettyJSON(data)
func (c *Context) SendPrettyJSON(value any, ctype ...string) error {
	return c.SendJSONIndent(value, " ", " ", ctype...)
}

// SendJSONIndent encodes the given value as indented (pretty-formatted) JSON
// using the specified prefix and indent strings. It writes the result to the response.
//
// This is useful for debugging or generating human-readable JSON output with custom formatting.
//
// It sets the Content-Type to "application/json" unless overridden by the optional ctype.
//
// Example:
//
//	err := c.SendJSONIndent(data, "", "  ") // pretty JSON with 2-space indent
//
//	err := c.SendJSONIndent(data, "--", ">>", "application/vnd.api+json")
func (c *Context) SendJSONIndent(value any, prefix, indent string, ctype ...string) error {
	contentType := "application/json"
	if len(ctype) > 0 {
		contentType = ctype[0]
	}
	c.SetContentType(contentType)

	bytes, err := sonic.MarshalIndent(value, prefix, indent)
	if err != nil {
		return NewHTTPError(StatusInternalServerError, "Failed to encode JSON: "+err.Error())
	}
	return c.SendBytes(bytes)
}

// SendSecureJSON encodes the value as JSON and adds a prefix to arrays
// to prevent JSON hijacking vulnerabilities in legacy browsers.
// The prefix (default: `while(1);`) is configured via Zeno().SecureJSONPrefix.
//
// Example:
//
//	return c.SendSecureJSON(data)
func (c *Context) SendSecureJSON(value any, ctype ...string) error {
	contentType := "application/json"
	if len(ctype) > 0 {
		contentType = ctype[0]
	}
	c.SetContentType(contentType)

	b, err := c.Zeno().JsonEncoder(value)
	if err != nil {
		return NewHTTPError(StatusInternalServerError,
			"Failed to encode JSON: "+err.Error())
	}

	//  If the payload starts with “[”, add the prefix
	trimmed := bytes.TrimLeft(b, " \t\r\n")
	if len(trimmed) > 0 && trimmed[0] == '[' {
		b = append([]byte(c.Zeno().SecureJSONPrefix), b...)
	}

	return c.SendBytes(b)
}

// SendXML encodes the given value as XML and writes it to the response.
//
// It sets the Content-Type to "application/xml; charset=utf-8" unless overridden
// via an optional parameter.
//
// Example:
//
//	type User struct {
//	    Name string `xml:"name"`
//	}
//
//	return c.SendXML(User{Name: "Alice"})
func (c *Context) SendXML(value any, ctype ...string) error {
	contentType := "application/xml; charset=utf-8"
	if len(ctype) > 0 {
		contentType = ctype[0]
	}
	c.SetContentType(contentType)

	b, err := c.Zeno().XmlEncoder(value)
	if err != nil {
		return NewHTTPError(StatusInternalServerError,
			"Failed to encode XML: "+err.Error())
	}

	return c.SendBytes(b)
}

// SendXMLIndent encodes the given value as indented (pretty-formatted) XML
// and writes it to the response.
//
// It sets the Content-Type to "application/xml; charset=utf-8" unless overridden.
// This is useful for development or human-readable output.
//
// Example:
//
//	type User struct {
//	    Name string `xml:"name"`
//	}
//
//	return c.SendXMLIndent(User{Name: "Alice"}, "", "  ")
func (c *Context) SendXMLIndent(value any, prefix, indent string, ctype ...string) error {
	contentType := "application/xml; charset=utf-8"
	if len(ctype) > 0 {
		contentType = ctype[0]
	}
	c.SetContentType(contentType)

	b, err := c.Zeno().XmlIndent(value, prefix, indent)
	if err != nil {
		return NewHTTPError(StatusInternalServerError,
			"Failed to encode XML: "+err.Error())
	}

	return c.SendBytes(b)
}

// BindXML decodes the request body as XML into the provided destination object.
//
// It returns a 400 error if the body is empty or if the XML is malformed.
//
// Example:
//
//	var user User
//	if err := c.BindXML(&user); err != nil {
//	    return err
//	}
func (c *Context) BindXML(out any) error {
	body := c.PostBody()
	if len(body) == 0 {
		return NewHTTPError(StatusBadRequest, "Request body is empty")
	}
	if err := c.Zeno().XmlDecoder(body, out); err != nil {
		return NewHTTPError(StatusBadRequest, "Invalid XML: "+err.Error())
	}
	return nil
}

// BindYAML reads the request body, decodes it as YAML, and stores the
// result in out.
//
// The decoder is the YamlDecoder configured on the parent *Zeno* instance.
// A 400 Bad Request error is returned if the body is empty or the data
// cannot be decoded.
//
// Example:
//
//	type User struct {
//	    Name string `yaml:"name"`
//	    Age  int    `yaml:"age"`
//	}
//
//	var u User
//	if err := c.BindYAML(&u); err != nil {
//	    c.Logger().Error(err)
//	    return
//	}
func (c *Context) BindYAML(out any) error {
	body := c.PostBody()
	if len(body) == 0 {
		return NewHTTPError(StatusBadRequest, "Request body is empty")
	}
	if err := c.Zeno().YamlDecoder(body, out); err != nil {
		return NewHTTPError(StatusBadRequest, "Invalid YAML: "+err.Error())
	}
	return nil
}

// SendYAML encodes v as YAML and writes it to the response.
//
// It sets the Content‑Type to "application/yaml; charset=utf‑8" unless a
// custom value is provided via ctype. A 500 Internal Server Error is
// returned if encoding fails.
//
// Example:
//
//	data := map[string]any{
//	    "status": "ok",
//	    "count":  42,
//	}
//	if err := c.SendYAML(data); err != nil {
//	    c.Logger().Error(err)
//	}
//
// To override the default content type:
//
//	if err := c.SendYAML(data, "text/x-yaml"); err != nil {
//	    ...
//	}
func (c *Context) SendYAML(v any, ctype ...string) error {
	contentType := "application/yaml; charset=utf-8"
	if len(ctype) > 0 {
		contentType = ctype[0]
	}
	c.SetContentType(contentType)
	bytes, err := c.Zeno().YamlEncoder(v)
	if err != nil {
		return NewHTTPError(StatusInternalServerError, "Failed to encode YAML: "+err.Error())
	}
	return c.SendBytes(bytes)
}

// BindTOML reads the request body, decodes it as TOML, and stores the
// result in out.
//
// The decoder is the TomlDecoder configured on the parent *Zeno* instance.
// A 400 Bad Request error is returned if the body is empty or the data
// cannot be decoded.
//
// Example:
//
//	type Config struct {
//	    Host string `toml:"host"`
//	    Port int    `toml:"port"`
//	}
//
//	var cfg Config
//	if err := c.BindTOML(&cfg); err != nil {
//	    c.Logger().Error(err)
//	    return
//	}
func (c *Context) BindTOML(out any) error {
	body := c.PostBody()
	if len(body) == 0 {
		return NewHTTPError(StatusBadRequest, "Request body is empty")
	}
	if err := c.Zeno().TomlDecoder(body, out); err != nil {
		return NewHTTPError(StatusBadRequest, "Invalid TOML: "+err.Error())
	}
	return nil
}

// SendTOML encodes v as TOML and writes it to the response.
//
// It sets the Content-Type to "application/toml; charset=utf-8" unless a
// custom value is provided via ctype. A 500 Internal Server Error is
// returned if encoding fails.
//
// Example:
//
//	settings := map[string]any{
//	    "mode": "production",
//	    "port": 8080,
//	}
//	if err := c.SendTOML(settings); err != nil {
//	    c.Logger().Error(err)
//	}
//
// To override the default content type:
//
//	if err := c.SendTOML(settings, "text/x-toml"); err != nil {
//	    ...
//	}
func (c *Context) SendTOML(v any, ctype ...string) error {
	contentType := "application/toml; charset=utf-8"
	if len(ctype) > 0 {
		contentType = ctype[0]
	}
	c.SetContentType(contentType)
	bytes, err := c.Zeno().TomlEncoder(v)
	if err != nil {
		return NewHTTPError(StatusInternalServerError, "Failed to encode TOML: "+err.Error())
	}
	return c.SendBytes(bytes)
}
