package zeno

import (
	"net/http"
	"sort"
	"strings"
	"sync"
	"unsafe"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/reuseport"
)

type Handler func(*Context) error

// Zeno is the main application struct for the framework.
// It stores routing trees, middleware, error handling logic,
// and manages request context pooling and execution.
type Zeno struct {
	Group // Root group for registering routes directly

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
}

// New creates and returns a new Zeno instance with default settings,
// initializes route trees, not found handlers, and context pooling.
func New() *Zeno {
	z := &Zeno{
		routes: make(map[string]*Route),
	}
	z.Group = *NewGroup("", z, nil)
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
	r.Group.Use(handlers...)
	r.notFoundHandlers = combineHandlers(r.handlers, r.notFound)
}

// Route returns a named route by name.
func (z *Zeno) Route(name string) *Route {
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
	c.init(ctx)
	c.handlers, c.pnames = z.find(z.toString(ctx.Method()), ctx.Path(), c.pvalues)
	if err := c.Next(); err != nil {
		z.ErrorHandler(c, err)
	}
	z.pool.Put(c)
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
	return DefaultNotFound
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

// Group returns a new RouteGroup whose path prefix is the current group’s
// prefix followed by prefix. Any handlers passed to Group are appended to the
// new group; if none are provided, the new group inherits the current group’s
// handlers.
//
// Example:
//
//	api := router.Group("/api", auth)
//	v1  := api.Group("/v1")           // -> prefix “/api/v1”, handlers {auth}
//
// prefix should begin with “/”; Group does not add a leading slash
// automatically.
func (r *Group) Group(prefix string, handlers ...Handler) *RouteGroup {
	if len(handlers) == 0 {
		handlers = make([]Handler, len(r.handlers))
		copy(handlers, r.handlers)
	}
	return NewGroup(r.prefix+prefix, r.router, handlers)
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
