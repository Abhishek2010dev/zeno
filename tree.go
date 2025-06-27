// Package zeno provides a high-performance HTTP router based on a radix tree
// structure with support for inline parameters, optional and wildcard tokens,
// and pattern-matching using regular expressions.
package zeno

import (
	"math"
	"regexp"
	"strings"
)

// tree represents a routing tree used for storing and matching routes.
// It consists of a root node and a counter for insertion order.
type tree struct {
	root  *node
	count int
}

// newTree creates and returns a new empty routing tree.
func newTree() *tree {
	return &tree{
		root: &node{
			static:    true,
			children:  make([]*node, 256),
			pchildren: []*node{},
			pindex:    -1,
			pnames:    []string{},
		},
	}
}

// Add inserts a new route into the tree with its corresponding handler chain.
// It returns the number of named parameters in the route.
func (t *tree) Add(key string, handlers []Handler) int {
	t.count++
	return t.root.add(key, handlers, t.count)
}

// Get matches the given path against the routing tree, writing parameter values
// to pvalues (which must be preallocated). It returns the matched handlers,
// parameter names, and the match's insertion order.
func (t *tree) Get(path string, pvalues []string) ([]Handler, []string) {
	d, names, _ := t.root.get(path, pvalues)
	return d, names
}

// node represents a single node in the routing tree. It may represent a static
// path segment, a parameterized segment, or a wildcard.
type node struct {
	static   bool           // true if the node is a static literal
	optional bool           // true if the param is optional (e.g., {id?})
	wildcard bool           // true if the param is a wildcard (e.g., {path*})
	key      string         // literal or token
	regex    *regexp.Regexp // compiled regex if the param has a pattern

	handlers []Handler // handler chain for the route
	order    int       // insertion order
	minOrder int       // minimum order of any handler in subtree

	children  []*node // static children (indexed by byte)
	pchildren []*node // parameter children

	pindex int      // index into pvalues
	pnames []string // ordered list of parameter names
}

// add inserts a route into the tree, splitting nodes as needed.
// It returns the number of parameters in the route.
func (n *node) add(key string, handlers []Handler, order int) int {
	matched := 0
	for matched < len(key) && matched < len(n.key) && key[matched] == n.key[matched] {
		matched++
	}

	if matched == len(n.key) && matched == len(key) {
		if n.handlers == nil {
			n.handlers = handlers
			n.order = order
		}
		return n.pindex + 1
	}

	if matched == len(n.key) {
		childKey := key[matched:]

		if lit := n.children[childKey[0]]; lit != nil {
			if pn := lit.add(childKey, handlers, order); pn >= 0 {
				return pn
			}
		}
		for _, pc := range n.pchildren {
			if pn := pc.add(childKey, handlers, order); pn >= 0 {
				return pn
			}
		}
		return n.addChild(childKey, handlers, order)
	}

	if matched == 0 || !n.static {
		return -1
	}

	rest := n.key[matched:]
	n1 := &node{
		static:    true,
		key:       rest,
		handlers:  n.handlers,
		order:     n.order,
		minOrder:  n.minOrder,
		children:  n.children,
		pchildren: n.pchildren,
		pindex:    n.pindex,
		pnames:    n.pnames,
		optional:  n.optional,
		wildcard:  n.wildcard,
		regex:     n.regex,
	}

	n.key = key[:matched]
	n.handlers = nil
	n.children = make([]*node, 256)
	n.pchildren = []*node{}
	n.children[rest[0]] = n1

	return n.add(key, handlers, order)
}

// addChild creates and attaches a new child node to the current node for the
// given path segment. It parses parameter tokens and patterns.
func (n *node) addChild(key string, handlers []Handler, order int) int {
	p0, p1 := -1, -1
	for i := 0; i < len(key); i++ {
		switch key[i] {
		case '{':
			p0 = i
		case '}':
			if p0 >= 0 {
				p1 = i
				i = len(key)
			}
		}
	}

	if p0 < 0 || p1 < 0 {
		// No parameter, treat as static
		lit := &node{
			static:    true,
			key:       key,
			minOrder:  order,
			children:  make([]*node, 256),
			pchildren: []*node{},
			pindex:    n.pindex,
			pnames:    n.pnames,
			handlers:  handlers,
			order:     order,
		}
		n.children[key[0]] = lit
		return lit.pindex + 1
	}

	// Static prefix before parameter
	if p0 > 0 {
		prefix := &node{
			static:    true,
			key:       key[:p0],
			minOrder:  order,
			children:  make([]*node, 256),
			pchildren: []*node{},
			pindex:    n.pindex,
			pnames:    n.pnames,
		}
		n.children[prefix.key[0]] = prefix
		return prefix.addChild(key[p0:], handlers, order)
	}

	token := key[p0 : p1+1]
	child := &node{
		static:    false,
		key:       token,
		minOrder:  order,
		children:  make([]*node, 256),
		pchildren: []*node{},
		pindex:    n.pindex,
		pnames:    n.pnames,
	}

	raw := token[1 : len(token)-1]

	// Handle wildcard
	if strings.HasPrefix(raw, "*") && len(raw) > 1 {
		raw = raw[1:] + "*"
	}

	pname, pattern := raw, ""
	if colon := strings.IndexByte(raw, ':'); colon >= 0 {
		pname, pattern = raw[:colon], raw[colon+1:]
	}

	if strings.HasSuffix(pname, "?") {
		child.optional = true
		pname = pname[:len(pname)-1]
	}

	if strings.HasSuffix(pname, "*") {
		child.wildcard = true
		pname = pname[:len(pname)-1]
		if p1+1 != len(key) {
			panic("routing: wildcard parameter must be terminal in pattern: " + key)
		}
	}

	if pattern != "" {
		child.regex = regexp.MustCompile("^" + pattern)
	}

	names := append([]string{}, child.pnames...)
	names = append(names, pname)
	child.pnames = names
	child.pindex = len(names) - 1

	n.pchildren = append(n.pchildren, child)

	if p1+1 == len(key) {
		child.handlers = handlers
		child.order = order
		return child.pindex + 1
	}
	return child.addChild(key[p1+1:], handlers, order)
}

// get attempts to match the given path against the current node and its children.
// It fills pvalues with any captured parameter values and returns the matched
// handler chain, parameter names, and order of match.
func (n *node) get(path string, pvalues []string) ([]Handler, []string, int) {
	bestOrder := math.MaxInt32
	var bestData []Handler
	var bestNames []string

repeat:
	if n.static {
		if !strings.HasPrefix(path, n.key) {
			return nil, nil, bestOrder
		}
		path = path[len(n.key):]
	} else if n.regex != nil {
		if len(path) == 0 && n.optional {
			pvalues[n.pindex] = ""
		} else if m := n.regex.FindStringIndex(path); m != nil {
			pvalues[n.pindex] = path[:m[1]]
			path = path[m[1]:]
		} else {
			return nil, nil, bestOrder
		}
	} else if n.wildcard {
		pvalues[n.pindex] = path
		path = ""
	} else {
		if len(path) == 0 {
			if n.optional {
				pvalues[n.pindex] = ""
			} else {
				return nil, nil, bestOrder
			}
		} else {
			idx := 0
			for idx < len(path) && path[idx] != '/' {
				if n.children[path[idx]] != nil {
					break
				}
				idx++
			}
			pvalues[n.pindex] = path[:idx]
			path = path[idx:]
		}
	}

	// Match static child
	if len(path) > 0 {
		if lit := n.children[path[0]]; lit != nil {
			if len(n.pchildren) == 0 {
				n = lit
				goto repeat
			}
			if d, names, o := lit.get(path, pvalues); d != nil && o < bestOrder {
				bestData, bestNames, bestOrder = d, names, o
			}
		}
	} else if n.handlers != nil {
		bestData, bestNames, bestOrder = n.handlers, n.pnames, n.order
	}

	// Try parametric children
	tmp := pvalues
	scratchAllocated := false
	for _, pc := range n.pchildren {
		if pc.minOrder >= bestOrder {
			continue
		}
		if bestData != nil && !scratchAllocated {
			tmp = make([]string, len(pvalues))
			scratchAllocated = true
		}
		if d, names, o := pc.get(path, tmp); d != nil && o < bestOrder {
			if scratchAllocated {
				copy(pvalues[pc.pindex:], tmp[pc.pindex:])
			}
			bestData, bestNames, bestOrder = d, names, o
		}
	}

	return bestData, bestNames, bestOrder
}
