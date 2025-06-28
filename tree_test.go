package zeno

import (
	"reflect"
	"testing"
)

func testHandler() Handler {
	return func(_ *Context) error {
		return nil
	}
}

func TestTree_AddAndGet(t *testing.T) {
	tree := newTree()
	pvalues := make([]string, 10)

	// Unique handlers to compare by pointer identity
	staticHandler := testHandler()
	paramHandler := testHandler()
	optionalHandler := testHandler()
	wildcardHandler := testHandler()
	regexHandler := testHandler()
	mixedHandler := testHandler()

	// Positive test cases
	tests := []struct {
		route      string
		path       string
		handler    Handler
		expectVals []string
	}{
		{"/about", "/about", staticHandler, nil},
		{"/user/{id}", "/user/123", paramHandler, []string{"123"}},
		{"/post/{id?}", "/post", optionalHandler, []string{""}},
		{"/post/{id?}", "/post/42", optionalHandler, []string{"42"}},
		{"/files/{path*}", "/files/docs/a.txt", wildcardHandler, []string{"docs/a.txt"}},
		{"/item/{slug:[a-z0-9\\-]+}", "/item/hello-123", regexHandler, []string{"hello-123"}},
		{"/page/{year}-{slug}", "/page/2022-zeno", mixedHandler, []string{"2022", "zeno"}},
	}

	for _, test := range tests {
		tree.Add([]byte(test.route), []Handler{test.handler})
	}

	for _, test := range tests {
		copy(pvalues, make([]string, len(pvalues)))
		handlers, pnames := tree.Get([]byte(test.path), pvalues)

		if test.handler == nil {
			if handlers != nil {
				t.Errorf("expected no match for path %q, got handler", test.path)
			}
			continue
		}

		if handlers == nil || len(handlers) == 0 {
			t.Errorf("expected handler for %q, got nil", test.path)
			continue
		}

		if reflect.ValueOf(handlers[0]).Pointer() != reflect.ValueOf(test.handler).Pointer() {
			t.Errorf("handler mismatch for path %q", test.path)
		}

		if test.expectVals != nil && !reflect.DeepEqual(pvalues[:len(test.expectVals)], test.expectVals) {
			t.Errorf("pvalues mismatch for path %q: expected %v, got %v", test.path, test.expectVals, pvalues[:len(test.expectVals)])
		}

		if len(pnames) != len(test.expectVals) {
			t.Errorf("pnames length mismatch for %q: expected %d, got %d", test.path, len(test.expectVals), len(pnames))
		}
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
