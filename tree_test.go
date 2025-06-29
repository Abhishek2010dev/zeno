package zeno

import (
	"testing"
)

func testHandler() Handler {
	return func(_ *Context) error {
		return nil
	}
}

func TestTree_NegativeMatches(t *testing.T) {
	tree := newTree()
	pvalues := make([]string, 10)

	// Add valid routes
	tree.Add([]byte("/user/{id}"), []Handler{testHandler()})
	tree.Add([]byte("/post/{id?}"), []Handler{testHandler()})
	tree.Add([]byte("/files/{path*}"), []Handler{testHandler()})
	tree.Add([]byte("/item/{slug:[a-z0-9\\-]+}"), []Handler{testHandler()})
	tree.Add([]byte("/page/{year}-{slug}"), []Handler{testHandler()})

	// Negative cases: these should NOT match any route
	negativePaths := []string{
		"/unknown",
		"/user",           // Missing param
		"/item/INVALID$$", // Fails regex
		"/page/2022zeno",  // Missing separator
		"/files",          // Missing wildcard (but valid only if optional)
		"/post/42/extra",  // Extra segment
		"/page/2022-",     // Missing slug
		"/item/hello_123", // Invalid regex character
	}

	for _, path := range negativePaths {
		copy(pvalues, make([]string, len(pvalues)))
		handlers, _ := tree.Get([]byte(path), pvalues)
		if handlers != nil {
			t.Errorf("expected no match for path %q, got handler", path)
		}
	}
}
