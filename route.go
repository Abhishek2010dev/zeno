// Package zeno provides a high-performance HTTP router and framework.
package zeno

import (
	"fmt"
	"net/url"
	"strings"
)

// Route represents a route definition, including its path, name,
// associated handlers, and belonging group.
type Route struct {
	group    *Group
	name     string
	path     string
	template string
}

// newRoute creates a new Route instance associated with the given group and path.
// It transforms wildcard patterns into regular expressions and builds a URL template.
//
// It also registers the route in the global Zeno routes map.
func newRoute(path string, group *Group) *Route {
	path = group.prefix + path
	name := path

	if strings.HasSuffix(path, "*") {
		path = path[:len(path)-1] + "{:.*}"
	}

	route := &Route{
		group:    group,
		name:     name,
		path:     path,
		template: buildURLTemplate(path),
	}
	route.group.zeno.routes[path] = route
	return route
}

// Name sets a custom name for the route and registers it using that name.
//
// Example:
//
//	r := newRoute("/user/{id}", group).Name("user.show")
func (r *Route) Name(name string) *Route {
	r.name = name
	r.group.zeno.routes[name] = r
	return r
}

// Get registers handlers for the GET HTTP method.
func (r *Route) Get(handlers ...Handler) *Route {
	return r.add("GET", handlers)
}

// Post registers handlers for the POST HTTP method.
func (r *Route) Post(handlers ...Handler) *Route {
	return r.add("POST", handlers)
}

// Put registers handlers for the PUT HTTP method.
func (r *Route) Put(handlers ...Handler) *Route {
	return r.add("PUT", handlers)
}

// Patch registers handlers for the PATCH HTTP method.
func (r *Route) Patch(handlers ...Handler) *Route {
	return r.add("PATCH", handlers)
}

// Delete registers handlers for the DELETE HTTP method.
func (r *Route) Delete(handlers ...Handler) *Route {
	return r.add("DELETE", handlers)
}

// Connect registers handlers for the CONNECT HTTP method.
func (r *Route) Connect(handlers ...Handler) *Route {
	return r.add("CONNECT", handlers)
}

// Head registers handlers for the HEAD HTTP method.
func (r *Route) Head(handlers ...Handler) *Route {
	return r.add("HEAD", handlers)
}

// Options registers handlers for the OPTIONS HTTP method.
func (r *Route) Options(handlers ...Handler) *Route {
	return r.add("OPTIONS", handlers)
}

// Trace registers handlers for the TRACE HTTP method.
func (r *Route) Trace(handlers ...Handler) *Route {
	return r.add("TRACE", handlers)
}

// To registers the same handlers for multiple comma-separated HTTP methods.
//
// Example:
//
//	r.To("GET,POST", handler)
func (r *Route) To(methods string, handlers ...Handler) *Route {
	for method := range strings.SplitSeq(methods, ",") {
		r.add(strings.TrimSpace(method), handlers)
	}
	return r
}

// add registers handlers for a single HTTP method and attaches route/middleware chain.
func (r *Route) add(method string, handlers []Handler) *Route {
	hh := combineHandlers(r.group.handlers, handlers)
	r.group.zeno.add(method, r.path, hh)
	return r
}

// buildURLTemplate creates a reusable path template by stripping regex
// suffixes from route parameters.
//
// Example:
// Input: "/users/{id:[0-9]+}/posts/{slug}"
// Output: "/users/{id}/posts/{slug}"
func buildURLTemplate(path string) string {
	template, start, end := "", -1, -1
	for i := 0; i < len(path); i++ {
		if path[i] == '{' && start < 0 {
			start = i
		} else if path[i] == '}' && start >= 0 {
			name := path[start+1 : i]
			for j := start + 1; j < i; j++ {
				if path[j] == ':' {
					name = path[start+1 : j]
					break
				}
			}
			template += path[end+1:start] + "{" + name + "}"
			end = i
			start = -1
		}
	}
	if end < 0 {
		template = path
	} else if end < len(path)-1 {
		template += path[end+1:]
	}
	return template
}

// URL generates a URL path from the route template and provided parameters.
//
// Example:
//
//	r := newRoute("/users/{id}", group).Name("user.show")
//	url := r.URL("id", 42) // => "/users/42"
func (r *Route) URL(pairs ...interface{}) (s string) {
	s = r.template
	for i := 0; i < len(pairs); i++ {
		name := fmt.Sprintf("{%v}", pairs[i])
		value := ""
		if i < len(pairs)-1 {
			value = url.QueryEscape(fmt.Sprint(pairs[i+1]))
		}
		s = strings.Replace(s, name, value, -1)
	}
	return
}

// combineHandlers merges group-level handlers with route-level handlers.
//
// The result is a flat handler chain with group handlers executed first.
func combineHandlers(h1 []Handler, h2 []Handler) []Handler {
	hh := make([]Handler, len(h1)+len(h2))
	copy(hh, h1)
	copy(hh[len(h1):], h2)
	return hh
}
