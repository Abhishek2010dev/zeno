package zeno

import (
	"bytes"
	"testing"

	"github.com/fxamacker/cbor/v2"
	"github.com/valyala/fasthttp"
)

// newTestContext creates a minimal *Context with a synthetic fasthttp.RequestCtx
// so individual helpers can be unit‐tested without starting a server.
func newTestContext(method, uri string, headers map[string]string, body []byte) (*Context, *fasthttp.RequestCtx) {
	req := fasthttp.AcquireRequest()

	req.Header.SetMethod(method)
	req.SetRequestURI(uri)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	if body != nil {
		req.SetBody(body)
	}

	ctxNative := &fasthttp.RequestCtx{}
	ctxNative.Init(req, nil, nil)

	z := New()

	c := &Context{
		RequestCtx: ctxNative,
		zeno:       z,
		index:      -1,
	}
	return c, ctxNative
}

func TestContext_Param(t *testing.T) {
	c, _ := newTestContext("GET", "/users/123", nil, nil)

	// Simulate router population
	c.pnames = []string{"id"}
	c.pvalues = []string{"123"}

	if got := c.Param("id"); got != "123" {
		t.Fatalf("Param(id) = %q; want %q", got, "123")
	}
	if got := c.Param("missing", "default"); got != "default" {
		t.Fatalf("Param missing default = %q; want %q", got, "default")
	}
}

func TestContext_Query(t *testing.T) {
	c, _ := newTestContext("GET", "/search?q=chatgpt&lang=en&lang=fr", nil, nil)

	if got := c.Query("q"); got != "chatgpt" {
		t.Fatalf("Query(q) = %q; want %q", got, "chatgpt")
	}
	if got := c.Query("none", "dft"); got != "dft" {
		t.Fatalf("Query default = %q; want %q", got, "dft")
	}

	langs := c.QueryArray("lang")
	if len(langs) != 2 || langs[0] != "en" || langs[1] != "fr" {
		t.Fatalf("QueryArray(lang) = %#v; want [en fr]", langs)
	}
}

func TestContext_Accepts(t *testing.T) {
	headers := map[string]string{
		"Accept": "application/json, text/html;q=0.8, */*;q=0.1",
	}
	c, _ := newTestContext("GET", "/", headers, nil)

	if got := c.Accepts("text/html", "application/json"); got != "application/json" {
		t.Fatalf("Accepts best = %q; want %q", got, "application/json")
	}
}

func TestContext_RealIP(t *testing.T) {
	headers := map[string]string{
		"X-Forwarded-For": "203.0.113.1, 70.41.3.18",
	}
	c, _ := newTestContext("GET", "/", headers, nil)

	if got := c.RealIP(); got != "203.0.113.1" {
		t.Fatalf("RealIP = %q; want %q", got, "203.0.113.1")
	}
}

func TestContext_SendString(t *testing.T) {
	c, native := newTestContext("GET", "/", nil, nil)

	if err := c.SendString("hello"); err != nil {
		t.Fatalf("SendString error = %v", err)
	}
	if got := string(native.Response.Body()); got != "hello" {
		t.Fatalf("response body = %q; want %q", got, "hello")
	}
}

func TestContext_Ranges(t *testing.T) {
	headers := map[string]string{
		"Range": "bytes=0-99,200-299",
	}
	c, _ := newTestContext("GET", "/", headers, nil)

	r, err := c.Ranges(500)
	if err != nil {
		t.Fatalf("Ranges error = %v", err)
	}
	if r == nil || r.Type != "bytes" || len(r.Ranges) != 2 {
		t.Fatalf("unexpected range result: %#v", r)
	}
	if r.Ranges[0].Start != 0 || r.Ranges[0].End != 99 {
		t.Errorf("first range got %#v; want {0 99}", r.Ranges[0])
	}
	if r.Ranges[1].Start != 200 || r.Ranges[1].End != 299 {
		t.Errorf("second range got %#v; want {200 299}", r.Ranges[1])
	}
}

type user struct {
	Name string `json:"name" xml:"name" yaml:"name" toml:"name" cbor:"name"`
	Age  int    `json:"age" xml:"age" yaml:"age" toml:"age" cbor:"age"`
}

func TestContext_JSON(t *testing.T) {
	input := []byte(`{"name":"Alice","age":30}`)
	c, native := newTestContext("POST", "/", map[string]string{
		"Content-Type": "application/json",
	}, input)

	var u user
	if err := c.BindJSON(&u); err != nil {
		t.Fatalf("BindJSON failed: %v", err)
	}
	if u.Name != "Alice" || u.Age != 30 {
		t.Fatalf("Parsed JSON incorrect: %+v", u)
	}

	if err := c.SendJSON(u); err != nil {
		t.Fatalf("SendJSON failed: %v", err)
	}
	if !bytes.Contains(native.Response.Body(), []byte(`"name":"Alice"`)) {
		t.Fatalf("response JSON = %s", native.Response.Body())
	}
}

func TestContext_XML(t *testing.T) {
	input := []byte(`<user><name>Alice</name><age>30</age></user>`)
	c, native := newTestContext("POST", "/", map[string]string{
		"Content-Type": "application/xml",
	}, input)

	var u user
	if err := c.BindXML(&u); err != nil {
		t.Fatalf("BindXML failed: %v", err)
	}

	if err := c.SendXML(u); err != nil {
		t.Fatalf("SendXML failed: %v", err)
	}
	if !bytes.Contains(native.Response.Body(), []byte(`<name>Alice</name>`)) {
		t.Fatalf("response XML = %s", native.Response.Body())
	}
}

func TestContext_YAML(t *testing.T) {
	input := []byte(`name: Alice
age: 30`)
	c, native := newTestContext("POST", "/", map[string]string{
		"Content-Type": "application/x-yaml",
	}, input)

	var u user
	if err := c.BindYAML(&u); err != nil {
		t.Fatalf("BindYAML failed: %v", err)
	}

	if err := c.SendYAML(u); err != nil {
		t.Fatalf("SendYAML failed: %v", err)
	}
	if !bytes.Contains(native.Response.Body(), []byte("name: Alice")) {
		t.Fatalf("response YAML = %s", native.Response.Body())
	}
}

func TestContext_CBOR(t *testing.T) {
	encoded, _ := cbor.Marshal(user{Name: "Alice", Age: 30})
	c, native := newTestContext("POST", "/", map[string]string{
		"Content-Type": "application/cbor",
	}, encoded)

	var u user
	if err := c.BindCBOR(&u); err != nil {
		t.Fatalf("BindCBOR failed: %v", err)
	}

	if err := c.SendCBOR(u); err != nil {
		t.Fatalf("SendCBOR failed: %v", err)
	}
	if len(native.Response.Body()) == 0 {
		t.Fatalf("response CBOR is empty")
	}
}

func TestContext_TOML(t *testing.T) {
	input := []byte(`name = "Alice"
age = 30`)

	c, native := newTestContext("POST", "/", map[string]string{
		"Content-Type": "application/toml",
	}, input)

	var u user
	if err := c.BindTOML(&u); err != nil {
		t.Fatalf("BindTOML failed: %v", err)
	}
	if u.Name != "Alice" || u.Age != 30 {
		t.Fatalf("TOML bind incorrect: %+v", u)
	}

	if err := c.SendTOML(u); err != nil {
		t.Fatalf("SendTOML failed: %v", err)
	}

	body := native.Response.Body()
	if !bytes.Contains(body, []byte(`name = "Alice"`)) &&
		!bytes.Contains(body, []byte(`name = 'Alice'`)) {
		t.Fatalf("response TOML missing name; got: %s", body)
	}
	if !bytes.Contains(body, []byte("age = 30")) {
		t.Fatalf("response TOML missing age; got: %s", body)
	}
}
