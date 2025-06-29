package zeno

import (
	"encoding/xml"
	"net/http"
	"sort"
	"strings"
	"sync"
	"unsafe"

	"github.com/bytedance/sonic"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/reuseport"
)

type Handler func(*Context) error

// Zeno is the main application struct for the framework.
// It stores routing trees, middleware, error handling logic,
// and manages request context pooling and execution.
type Zeno struct {
	RouteGroup // Root group for registering routes directly

	// Routing trees for each HTTP method
	getTree     *tree
	headTree    *tree
	postTree    *tree
	putTree     *tree
	patchTree   *tree
	deletedTree *tree
	connectTree *tree
	optionsTree *tree
	traceTree   *tree

	// Request context pooling for performance
	pool sync.Pool

	// Max number of parameters used across all routes
	maxParams int

	// Handlers executed when no route matches
	notFound         []Handler
	notFoundHandlers []Handler

	// Named route registry
	routes map[string]*Route

	// Unsafe byte slice to string conversion
	toString func(v []byte) string

	// Custom error handler
	ErrorHandler func(*Context, error) error

	// Use SO_REUSEPORT for multiple listeners on same port
	useReusePort bool

	// JsonDecoder is the default function used to decode a JSON payload
	// from the request body. It should unmarshal the byte slice into
	// the target Go value. A typical implementation uses json.Unmarshal
	// or a high-performance alternative such as sonic or jsoniter.
	JsonDecoder DecoderFunc

	// JsonEncoder is the default function used to encode a Go value into
	// JSON format. It should return the marshaled byte slice that can be
	// directly written to the response. Set the "Content-Type" to
	// "application/json" before sending the bytes.
	JsonEncoder EncoderFunc

	// JsonIndent is an optional function used to pretty-print JSON output.
	// It takes a Go value, prefix, and indent string to format the output
	// for better readability. Typically wraps json.MarshalIndent or similar.
	JsonIndent IndentFunc

	// SecureJSONPrefix is a string prepended to all JSON responses
	// to prevent JSON Hijacking attacks. Common value: "while(1);"
	// If set, all JSON responses will begin with this prefix.
	SecureJSONPrefix string

	// XmlDecoder is the default function used to decode an XML payload
	// from the request body. It should unmarshal the byte slice into
	// the target Go value. Typically wraps encoding/xml.Unmarshal or
	// a faster XML decoder.
	XmlDecoder DecoderFunc

	// XmlEncoder is the default function used to encode a Go value into
	// XML format. It should return the marshaled byte slice that can be
	// written directly to the response. You should set the
	// "Content-Type" to "application/xml" or "text/xml" before writing.
	XmlEncoder EncoderFunc

	// XmlIndent is an optional function used to pretty-print XML output.
	// It takes a Go value, prefix, and indent string to format the output.
	// Usually wraps xml.MarshalIndent or any compatible alternative.
	XmlIndent IndentFunc
}

// New creates and returns a new Zeno instance with default settings,
// initializes route trees, not found handlers, and context pooling.
func New() *Zeno {
	z := &Zeno{
		routes:           make(map[string]*Route),
		JsonDecoder:      sonic.Unmarshal,
		JsonEncoder:      sonic.Marshal,
		JsonIndent:       sonic.MarshalIndent,
		XmlEncoder:       xml.Marshal,
		XmlDecoder:       xml.Unmarshal,
		XmlIndent:        xml.MarshalIndent,
		SecureJSONPrefix: "while(1);",
	}
	z.RouteGroup = *NewRouteGroup("", z, nil)
	z.pool.New = func() interface{} {
		return &Context{
			pvalues: make([]string, z.maxParams),
			zeno:    z,
		}
	}
	z.toString = func(b []byte) string {
		return *(*string)(unsafe.Pointer(&b))
	}
	z.NotFound(MethodNotAllowedHandler, NotFoundHandler)
	z.ErrorHandler = func(c *Context, err error) error {
		if httpErr, ok := err.(HTTPError); ok {
			return c.Status(httpErr.StatusCode()).SendString(httpErr.Error())
		}
		return c.Status(StatusInternalServerError).SendString("Internal Server Error")
	}
	return z
}

// Use appends the specified handlers to the router and shares them with all routes.
func (r *Zeno) Use(handlers ...Handler) {
	r.RouteGroup.Use(handlers...)
	r.notFoundHandlers = combineHandlers(r.handlers, r.notFound)
}

// Route returns a named route by name.
func (z *Zeno) GetRoute(name string) *Route {
	return z.routes[name]
}

// NotFound sets the handler(s) to be used when no route is matched.
// The final notFound handler chain includes global middleware.
func (r *Zeno) NotFound(handlers ...Handler) {
	r.notFound = handlers
	r.notFoundHandlers = combineHandlers(r.handlers, r.notFound)
}

// find attempts to locate a handler chain for the given method and path.
// If no match is found, the notFound handler is returned.
func (z *Zeno) find(method string, path []byte, pvalues []string) ([]Handler, []string) {
	t := z.treeForMethod(method)
	if t != nil {
		if h, pnames := t.Get(path, pvalues); h != nil {
			return h, pnames
		}
	}
	return z.notFoundHandlers, nil
}

// findAllowedMethods returns a set of allowed HTTP methods for a given path.
// Useful for generating Allow headers when responding with 405 errors.
func (z *Zeno) findAllowedMethods(path []byte) map[string]bool {
	methods := make(map[string]bool)
	pvalues := make([]string, z.maxParams)

	check := func(method string, s *tree) {
		if s != nil {
			if h, _ := s.Get(path, pvalues); h != nil {
				methods[method] = true
			}
		}
	}

	check(MethodGet, z.getTree)
	check(MethodHead, z.headTree)
	check(MethodPost, z.postTree)
	check(MethodPut, z.putTree)
	check(MethodPatch, z.patchTree)
	check(MethodDelete, z.deletedTree)
	check(MethodConnect, z.connectTree)
	check(MethodOptions, z.optionsTree)
	check(MethodTrace, z.traceTree)

	return methods
}

// HandleRequest is the main request entry point for fasthttp.
// It acquires a context from the pool, performs route matching,
// executes the handler chain, and handles any returned errors.
func (z *Zeno) HandleRequest(ctx *fasthttp.RequestCtx) {
	c := z.pool.Get().(*Context)
	defer z.pool.Put(c)

	c.init(ctx)
	c.handlers, c.pnames = z.find(z.toString(ctx.Method()), ctx.Path(), c.pvalues)

	if err := c.Next(); err != nil {
		// Call error handler if set
		if z.ErrorHandler != nil {
			if handleErr := z.ErrorHandler(c, err); handleErr != nil {
				c.SendStatusCode(StatusInternalServerError)
			}
		} else {
			// Fallback to default error response if no error handler is defined
			c.SendStatusCode(StatusInternalServerError)
		}
	}
}

// add registers a route in the routing tree for the given HTTP method.
// It updates maxParams if the route uses more parameters than seen so far.
func (z *Zeno) add(method, path string, handlers []Handler) {
	tree := z.treeForMethod(method)
	if tree == nil {
		tree = newTree()
		z.setTreeForMethod(method, tree)
	}
	if n := tree.Add([]byte(path), handlers); n > z.maxParams {
		z.maxParams = n
	}
}

// treeForMethod returns the routing tree corresponding to an HTTP method.
func (z *Zeno) treeForMethod(method string) *tree {
	switch method {
	case MethodGet:
		return z.getTree
	case MethodHead:
		return z.headTree
	case MethodPost:
		return z.postTree
	case MethodPut:
		return z.putTree
	case MethodPatch:
		return z.patchTree
	case MethodDelete:
		return z.deletedTree
	case MethodConnect:
		return z.connectTree
	case MethodOptions:
		return z.optionsTree
	case MethodTrace:
		return z.traceTree
	default:
		return nil
	}
}

// setTreeForMethod sets the routing tree for the given HTTP method.
func (z *Zeno) setTreeForMethod(method string, t *tree) {
	switch method {
	case MethodGet:
		z.getTree = t
	case MethodHead:
		z.headTree = t
	case MethodPost:
		z.postTree = t
	case MethodPut:
		z.putTree = t
	case MethodPatch:
		z.patchTree = t
	case MethodDelete:
		z.deletedTree = t
	case MethodConnect:
		z.connectTree = t
	case MethodOptions:
		z.optionsTree = t
	case MethodTrace:
		z.traceTree = t
	}
}

// NotFoundHandler is the default fallback handler that returns 404.
func NotFoundHandler(*Context) error {
	return ErrNotFound
}

// MethodNotAllowedHandler builds and sets the "Allow" header when
// a route exists for the path but not for the method. If the request
// method is not OPTIONS, it returns 405 Method Not Allowed.
func MethodNotAllowedHandler(c *Context) error {
	methods := c.Zeno().findAllowedMethods(c.RequestCtx.Path())
	if len(methods) == 0 {
		return nil
	}
	methods["OPTIONS"] = true
	ms := make([]string, 0, len(methods))
	for m := range methods {
		ms = append(ms, m)
	}
	sort.Strings(ms)
	c.RequestCtx.Response.Header.Set("Allow", strings.Join(ms, ", "))
	if string(c.RequestCtx.Method()) != "OPTIONS" {
		c.RequestCtx.Response.SetStatusCode(http.StatusMethodNotAllowed)
	}
	c.Abort()
	return nil
}

// Run starts the HTTP server on the given address using fasthttp.
// If useReusePort is true, it uses SO_REUSEPORT for load balancing across processes.
func (z *Zeno) Run(addr string) error {
	if z.useReusePort {
		ln, err := reuseport.Listen("tcp4", addr)
		if err != nil {
			return err
		}
		return fasthttp.Serve(ln, z.HandleRequest)
	}
	return fasthttp.ListenAndServe(addr, z.HandleRequest)
}
