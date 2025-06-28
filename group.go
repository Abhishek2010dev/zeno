package zeno

import "strings"

// Group represents a collection of routes with a common prefix and shared middleware handlers.
// It allows organizing routes into subgroups for modular design.
type Group struct {
	prefix   string    // Common path prefix for all routes in the group
	zeno     *Zeno     // Reference to the parent Zeno instance
	handlers []Handler // Middleware handlers applied to all routes in the group
}

// NewGroup creates and returns a new route group with the given path prefix,
// associated Zeno instance, and optional middleware handlers.
//
// Example:
//
//	api := NewGroup("/api", app, []Handler{authMiddleware})
func NewGroup(prefix string, zeno *Zeno, handlers []Handler) *Group {
	return &Group{
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
func (r *Group) Get(path string, handlers ...Handler) *Route {
	return newRoute(path, r).Get(handlers...)
}

// Post registers a new route in the group for the POST HTTP method.
//
// Example:
//
//	g.Post("/submit", submitHandler)
func (r *Group) Post(path string, handlers ...Handler) *Route {
	return newRoute(path, r).Post(handlers...)
}

// Put registers a new route in the group for the PUT HTTP method.
func (r *Group) Put(path string, handlers ...Handler) *Route {
	return newRoute(path, r).Put(handlers...)
}

// Patch registers a new route in the group for the PATCH HTTP method.
func (r *Group) Patch(path string, handlers ...Handler) *Route {
	return newRoute(path, r).Patch(handlers...)
}

// Delete registers a new route in the group for the DELETE HTTP method.
func (r *Group) Delete(path string, handlers ...Handler) *Route {
	return newRoute(path, r).Delete(handlers...)
}

// Connect registers a new route in the group for the CONNECT HTTP method.
func (r *Group) Connect(path string, handlers ...Handler) *Route {
	return newRoute(path, r).Connect(handlers...)
}

// Head registers a new route in the group for the HEAD HTTP method.
func (r *Group) Head(path string, handlers ...Handler) *Route {
	return newRoute(path, r).Head(handlers...)
}

// Options registers a new route in the group for the OPTIONS HTTP method.
func (r *Group) Options(path string, handlers ...Handler) *Route {
	return newRoute(path, r).Options(handlers...)
}

// Trace registers a new route in the group for the TRACE HTTP method.
func (r *Group) Trace(path string, handlers ...Handler) *Route {
	return newRoute(path, r).Trace(handlers...)
}

// To registers a new route in the group for multiple HTTP methods,
// specified as a comma-separated string (e.g. "GET,POST").
//
// Example:
//
//	g.To("GET,POST", "/users", usersHandler)
func (r *Group) To(methods, path string, handlers ...Handler) *Route {
	route := newRoute(path, r)
	for method := range strings.SplitSeq(methods, ",") {
		route.add(method, handlers)
	}
	return route
}
