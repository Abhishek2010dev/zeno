// Package zeno provides a high-performance HTTP router based on a radix tree
// structure with support for inline parameters, optional and wildcard tokens,
// and pattern-matching using regular expressions.
package zeno

import (
	"bytes"
	"math"
	"regexp"
)

// tree represents a routing tree used to store and match HTTP routes.
// Each tree corresponds to a specific HTTP method (e.g. GET, POST).
type tree struct {
	root  *node // root node of the routing tree
	count int   // total number of routes inserted
}

// newTree creates and returns a new empty routing tree with an initialized root node.
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

// Add inserts a new route path and its associated handler chain into the tree.
// The key must be a byte slice representing the route path (e.g., []byte("/users/{id}").
// It returns the number of named parameters in the route.
func (t *tree) Add(key []byte, handlers []Handler) int {
	t.count++
	return t.root.add(key, handlers, t.count)
}

// Get attempts to match the given path against the routing tree.
// It fills the provided pvalues slice with extracted parameter values.
// It returns the matched handler chain, ordered list of parameter names, and insertion order.
func (t *tree) Get(path []byte, pvalues []string) ([]Handler, []string) {
	d, names, _ := t.root.get(path, pvalues)
	return d, names
}

// node represents a single node in the radix tree.
// Nodes may represent static paths or parameterized segments like {id}, {slug:.*}, {file*}, or {name?}.
type node struct {
	static   bool           // true if the node is a static (literal) segment
	optional bool           // true if the parameter is optional
	wildcard bool           // true if the parameter captures the rest of the path
	key      []byte         // the literal or token segment of the path
	regex    *regexp.Regexp // compiled regex for pattern-matched parameters

	handlers []Handler // list of handlers to be called on match
	order    int       // insertion order of the route
	minOrder int       // minimum order of any handler in subtree (used for prioritization)

	children  []*node // static children (indexed by byte)
	pchildren []*node // parameterized children

	pindex int      // index of the parameter in the pvalues slice
	pnames []string // list of parameter names in order of appearance
}

// add inserts a new route key into the radix tree recursively.
// It returns the number of parameters added to the route.
func (n *node) add(key []byte, handlers []Handler, order int) int {
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

	// Split the current node into prefix and suffix
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

// addChild creates and attaches a new child node for the given path segment.
// It parses parameters (e.g. {id}, {slug:.*}, {name?}) and wildcards (e.g. {file*}).
func (n *node) addChild(key []byte, handlers []Handler, order int) int {
	p0, p1 := -1, -1
	for i := 0; i < len(key); i++ {
		switch key[i] {
		case '{':
			p0 = i
		case '}':
			if p0 >= 0 {
				p1 = i
				i = len(key) // break
			}
		}
	}

	if p0 < 0 || p1 < 0 {
		// Static literal
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

	if p0 > 0 {
		// Static prefix before parameter
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

	// Parameter token
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
	if len(raw) > 1 && raw[0] == '*' {
		raw = append(raw[1:], '*')
	}

	colon := bytes.IndexByte(raw, ':')
	pname, pattern := raw, []byte{}
	if colon >= 0 {
		pname = raw[:colon]
		pattern = raw[colon+1:]
	}

	if len(pname) > 0 && pname[len(pname)-1] == '?' {
		child.optional = true
		pname = pname[:len(pname)-1]
	}

	if len(pname) > 0 && pname[len(pname)-1] == '*' {
		child.wildcard = true
		pname = pname[:len(pname)-1]
		if p1+1 != len(key) {
			panic("routing: wildcard parameter must be terminal in pattern: " + string(key))
		}
	}

	if len(pattern) > 0 {
		child.regex = regexp.MustCompile("^" + string(pattern))
	}

	names := append([]string{}, child.pnames...)
	names = append(names, string(pname))
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

// get attempts to match a path against this node and its children recursively.
// It fills pvalues with captured parameter values and returns the matched
// handler chain, parameter names, and match insertion order.
func (n *node) get(path []byte, pvalues []string) ([]Handler, []string, int) {
	bestOrder := math.MaxInt32
	var bestData []Handler
	var bestNames []string

repeat:
	if n.static {
		if !bytes.HasPrefix(path, n.key) {
			return nil, nil, bestOrder
		}
		path = path[len(n.key):]
	} else if n.regex != nil {
		if len(path) == 0 && n.optional {
			pvalues[n.pindex] = ""
		} else if m := n.regex.FindIndex(path); m != nil {
			pvalues[n.pindex] = string(path[:m[1]])
			path = path[m[1]:]
		} else {
			return nil, nil, bestOrder
		}
	} else if n.wildcard {
		pvalues[n.pindex] = string(path)
		path = nil
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
			pvalues[n.pindex] = string(path[:idx])
			path = path[idx:]
		}
	}

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

	tmp := pvalues
	scratch := false
	for _, pc := range n.pchildren {
		if pc.minOrder >= bestOrder {
			continue
		}
		if bestData != nil && !scratch {
			tmp = make([]string, len(pvalues))
			scratch = true
		}
		if d, names, o := pc.get(path, tmp); d != nil && o < bestOrder {
			if scratch {
				copy(pvalues[pc.pindex:], tmp[pc.pindex:])
			}
			bestData, bestNames, bestOrder = d, names, o
		}
	}

	return bestData, bestNames, bestOrder
}
