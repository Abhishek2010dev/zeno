package zeno

import "strings"

// RouteGroup represents a collection of routes with a common prefix and shared middleware handlers.
// It allows organizing routes into subgroups for modular design.
type RouteGroup struct {
	prefix   string    // Common path prefix for all routes in the group
	zeno     *Zeno     // Reference to the parent Zeno instance
	handlers []Handler // Middleware handlers applied to all routes in the group
}

// NewRouteGroup creates and returns a new route group with the given path prefix,
// associated Zeno instance, and optional middleware handlers.
//
// Example:
//
//	api := NewRouteGroup("/api", app, []Handler{authMiddleware})
func NewRouteGroup(prefix string, zeno *Zeno, handlers []Handler) *RouteGroup {
	return &RouteGroup{
		prefix:   prefix,
		zeno:     zeno,
		handlers: handlers,
	}
}

// Get registers a new route in the group for the GET HTTP method.
//
// Example:
//
//	g.Get("/ping", pingHandler)
func (r *RouteGroup) Get(path string, handlers ...Handler) *Route {
	return newRoute(path, r).Get(handlers...)
}

// Post registers a new route in the group for the POST HTTP method.
//
// Example:
//
//	g.Post("/submit", submitHandler)
func (r *RouteGroup) Post(path string, handlers ...Handler) *Route {
	return newRoute(path, r).Post(handlers...)
}

// Put registers a new route in the group for the PUT HTTP method.
func (r *RouteGroup) Put(path string, handlers ...Handler) *Route {
	return newRoute(path, r).Put(handlers...)
}

// Patch registers a new route in the group for the PATCH HTTP method.
func (r *RouteGroup) Patch(path string, handlers ...Handler) *Route {
	return newRoute(path, r).Patch(handlers...)
}

// Delete registers a new route in the group for the DELETE HTTP method.
func (r *RouteGroup) Delete(path string, handlers ...Handler) *Route {
	return newRoute(path, r).Delete(handlers...)
}

// Connect registers a new route in the group for the CONNECT HTTP method.
func (r *RouteGroup) Connect(path string, handlers ...Handler) *Route {
	return newRoute(path, r).Connect(handlers...)
}

// Head registers a new route in the group for the HEAD HTTP method.
func (r *RouteGroup) Head(path string, handlers ...Handler) *Route {
	return newRoute(path, r).Head(handlers...)
}

// Options registers a new route in the group for the OPTIONS HTTP method.
func (r *RouteGroup) Options(path string, handlers ...Handler) *Route {
	return newRoute(path, r).Options(handlers...)
}

// Trace registers a new route in the group for the TRACE HTTP method.
func (r *RouteGroup) Trace(path string, handlers ...Handler) *Route {
	return newRoute(path, r).Trace(handlers...)
}

// To registers a new route in the group for multiple HTTP methods,
// specified as a comma-separated string (e.g. "GET,POST").
//
// Example:
//
//	g.To("GET,POST", "/users", usersHandler)
func (r *RouteGroup) To(methods, path string, handlers ...Handler) *Route {
	route := newRoute(path, r)
	for method := range strings.SplitSeq(methods, ",") {
		route.add(method, handlers)
	}
	return route
}

// Use registers one or multiple handlers to the current route group.
// These handlers will be shared by all routes belong to this group and its subgroups.
func (r *RouteGroup) Use(handlers ...Handler) {
	r.handlers = append(r.handlers, handlers...)
}

// RouteGroup returns a new RouteRouteGroup whose path prefix is the current group’s
// prefix followed by prefix. Any handlers passed to RouteGroup are appended to the
// new group; if none are provided, the new group inherits the current group’s
// handlers.
//
// Example:
//
//	api := router.RouteGroup("/api", auth)
//	v1  := api.RouteGroup("/v1")           // -> prefix “/api/v1”, handlers {auth}
//
// prefix should begin with “/”; RouteGroup does not add a leading slash
// automatically.
func (r *RouteGroup) Group(prefix string, handlers ...Handler) *RouteGroup {
	if len(handlers) == 0 {
		handlers = make([]Handler, len(r.handlers))
		copy(handlers, r.handlers)
	}
	return NewRouteGroup(r.prefix+prefix, r.zeno, handlers)
}

// Route creates a new sub-route group with the given path prefix and optional
// handlers. It then executes the provided function with the new group.
//
// This enables nesting of routes in a structured way, similar to Chi:
//
//	r.Route("/api", func(r *RouteGroup) {
//	    r.Get("/users", listUsers)
//	})
func (r *RouteGroup) Route(prefix string, fn func(*RouteGroup), handlers ...Handler) {
	g := r.Group(r.prefix+prefix, handlers...)
	fn(g)
}
