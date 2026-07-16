package web

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestRequest_Application(t *testing.T) {
	app := Default()
	req := httptest.NewRequest(http.MethodGet, "http://example.test/", nil)
	r := newRequest(app, req)
	if r.Application() != app {
		t.Fatalf("expected Application() to return the original application instance")
	}

	r = newRequest(nil, req)
	if r.Application() != nil {
		t.Fatalf("expected Application() to return nil when no application is associated")
	}
}

func TestRequest_Body(t *testing.T) {
	body := bytes.NewReader([]byte("hello world"))
	req := httptest.NewRequest(http.MethodPost, "http://example.test/", body)

	r := newRequest(nil, req)

	rc, err := r.Body()
	if err != nil {
		t.Fatalf("Body error: %v", err)
	}
	b, _ := io.ReadAll(rc)
	if string(b) != "hello world" {
		t.Fatalf("expected body 'hello world', got %q", string(b))
	}
}

func TestRequest_Body_LimitExceeded(t *testing.T) {
	body := strings.NewReader(strings.Repeat("x", 100))
	req := httptest.NewRequest(http.MethodPost, "http://example.test/", body)

	app := New(WithMaxRequestBodyBytes(50))
	r := newRequest(app, req)

	rc, err := r.Body()
	if err != nil {
		t.Fatalf("Body error: %v", err)
	}

	_, err = io.ReadAll(rc)
	if err == nil {
		t.Fatalf("expected MaxBytesError when body exceeds limit, got nil")
	}
	var maxBytesErr *http.MaxBytesError
	if !errors.As(err, &maxBytesErr) {
		t.Fatalf("expected *http.MaxBytesError, got %T: %v", err, err)
	}
	if maxBytesErr.Limit != 50 {
		t.Fatalf("expected limit 50, got %d", maxBytesErr.Limit)
	}
}

func TestRequest_Body_Unlimited(t *testing.T) {
	body := strings.NewReader(strings.Repeat("x", 100))
	req := httptest.NewRequest(http.MethodPost, "http://example.test/", body)

	app := New(WithMaxRequestBodyBytes(MaxRequestBodyBytesUnlimited))
	r := newRequest(app, req)

	rc, err := r.Body()
	if err != nil {
		t.Fatalf("Body error: %v", err)
	}

	data, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll error: %v", err)
	}
	if len(data) != 100 {
		t.Fatalf("expected 100 bytes, got %d", len(data))
	}
}

func TestRequest_Body_Idempotent(t *testing.T) {
	body := bytes.NewReader([]byte("hello"))
	req := httptest.NewRequest(http.MethodPost, "http://example.test/", body)
	r := newRequest(nil, req)

	rc1, _ := r.Body()
	rc2, _ := r.Body()

	if rc1 != rc2 {
		t.Fatalf("expected same reader from multiple Body() calls (sync.Once), got different")
	}

	data, _ := io.ReadAll(rc1)
	if string(data) != "hello" {
		t.Fatalf("expected 'hello', got %q", string(data))
	}
}

func TestRequest_ClientIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.test/", nil)
	r := newRequest(nil, req)

	ip := r.ClientIP()
	if ip != "192.0.2.1" {
		t.Fatalf("expected ClientIP '192.0.2.1', got %v", ip)
	}
}

func TestRequest_ContentLength(t *testing.T) {
	body := bytes.NewReader([]byte("data"))
	req := httptest.NewRequest(http.MethodPost, "http://example.test/", body)
	r := newRequest(nil, req)

	expectLen := int64(len("data"))
	if r.ContentLength() != expectLen {
		t.Fatalf("expected ContentLength %d, got %d", expectLen, r.ContentLength())
	}
}

func TestRequest_Cookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.test/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "abc123"})

	r := newRequest(nil, req)

	cookie, err := r.Cookie("session")
	if err != nil {
		t.Fatalf("Cookie error: %v", err)
	}
	if cookie.Value != "abc123" {
		t.Fatalf("expected cookie value 'abc123', got %q", cookie.Value)
	}
}

func TestRequest_Cookies(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.test/", nil)
	req.AddCookie(&http.Cookie{Name: "a", Value: "1"})
	req.AddCookie(&http.Cookie{Name: "b", Value: "2"})

	r := newRequest(nil, req)

	cookies := r.Cookies()
	if len(cookies) != 2 {
		t.Fatalf("expected 2 cookies, got %d", len(cookies))
	}
	if cookies[0].Name != "a" || cookies[0].Value != "1" {
		t.Fatalf("unexpected first cookie: %v", cookies[0])
	}
	if cookies[1].Name != "b" || cookies[1].Value != "2" {
		t.Fatalf("unexpected second cookie: %v", cookies[1])
	}
}

func TestRequest_Context(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.test/", nil)
	ctx := req.Context()

	r := newRequest(nil, req)

	if !reflect.DeepEqual(r.Context(), ctx) {
		t.Fatalf("expected Context to match request context")
	}
}

func TestRequest_Headers(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.test/", nil)
	req.Header.Add("X-Test", "v1")
	req.Header.Add("X-Test", "v2")
	req.Header.Set("X-Another", "v3")

	r := newRequest(nil, req)

	if r.Header("X-Test") != "v1" {
		t.Fatalf("expected Header 'X-Test' to be 'v1', got %v", r.Header("X-Test"))
	}
	if !reflect.DeepEqual(r.HeaderValues("X-Test"), []string{"v1", "v2"}) {
		t.Fatalf("expected HeaderValues 'X-Test' to be ['v1', 'v2'], got %v", r.HeaderValues("X-Test"))
	}
	if r.Header("X-Another") != "v3" {
		t.Fatalf("expected Header 'X-Another' to be 'v3', got %v", r.Header("X-Another"))
	}
	if r.Header("Non-Existent") != "" {
		t.Fatalf("expected non-existent header to return empty string, got %v", r.Header("Non-Existent"))
	}
	if len(r.HeaderValues("Non-Existent")) != 0 {
		t.Fatalf("expected non-existent header values to return empty slice, got %v", r.HeaderValues("Non-Existent"))
	}

	headers := r.Headers()
	if headers.Get("X-Test") != "v1" {
		t.Fatalf("expected Headers Get 'X-Test' to be 'v1', got %v", headers.Get("X-Test"))
	}
	if !reflect.DeepEqual(headers.Values("X-Test"), []string{"v1", "v2"}) {
		t.Fatalf("expected Headers Values 'X-Test' to be ['v1', 'v2'], got %v", headers.Values("X-Test"))
	}
	if headers.Get("X-Another") != "v3" {
		t.Fatalf("expected Headers Get 'X-Another' to be 'v3', got %v", headers.Get("X-Another"))
	}
}

func TestRequest_Method_Protocol_Path(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "http://example.test/test", nil)
	r := newRequest(nil, req)

	if r.Method() != http.MethodPost {
		t.Fatalf("expected method POST, got %v", r.Method())
	}
	if r.Protocol() != "HTTP/1.1" {
		t.Fatalf("expected protocol HTTP/1.1, got %v", r.Protocol())
	}
	if r.Path() != "/test" {
		t.Fatalf("expected path '/test', got %v", r.Path())
	}
}

func TestRequest_PathValue(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.test/", nil)
	r := newRequest(nil, req)

	val := r.PathValue("id")
	if val != "" {
		t.Fatalf("expected empty string for non-existent path parameter, got %v", val)
	}

	r.SetPathValue("id", "123")
	val = r.PathValue("id")
	if val != "123" {
		t.Fatalf("expected path value '123', got %v", val)
	}
}

func TestRequest_Resource(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.test/", nil)
	r := newRequest(nil, req)

	if r.Resource() != "" {
		t.Fatalf("expected empty string for resource, got %v", r.Resource())
	}
	r.SetResource("/test")
	if r.Resource() != "/test" {
		t.Fatalf("expected resource '/test', got %v", r.Resource())
	}
}

func TestRequest_Queries(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.test/search?q=go&q=web&lang=en", nil)
	r := newRequest(nil, req)

	if r.Query("q") != "go" {
		t.Fatalf("expected query 'q' to be 'go', got %v", r.Query("q"))
	}
	if r.Query("lang") != "en" {
		t.Fatalf("expected query 'lang' to be 'en', got %v", r.Query("lang"))
	}
	if r.Query("nonexistent") != "" {
		t.Fatalf("expected non-existent query parameter to return empty string, got %v", r.Query("nonexistent"))
	}

	if !reflect.DeepEqual(r.QueryValues("q"), []string{"go", "web"}) {
		t.Fatalf("expected query values 'q' to be ['go', 'web'], got %v", r.QueryValues("q"))
	}
	if !reflect.DeepEqual(r.QueryValues("lang"), []string{"en"}) {
		t.Fatalf("expected query values 'lang' to be ['en'], got %v", r.QueryValues("lang"))
	}
	if len(r.QueryValues("nonexistent")) != 0 {
		t.Fatalf("expected non-existent query parameter values to return empty slice, got %v", r.QueryValues("nonexistent"))
	}

	queries := r.Queries()
	if queries.Get("q") != "go" {
		t.Fatalf("expected Queries Get 'q' to be 'go', got %v", queries.Get("q"))
	}
	if !reflect.DeepEqual(queries["q"], []string{"go", "web"}) {
		t.Fatalf("expected Queries Values 'q' to be ['go', 'web'], got %v", queries["q"])
	}
	if queries.Get("lang") != "en" {
		t.Fatalf("expected Queries Get 'lang' to be 'en', got %v", queries.Get("lang"))
	}
	if !reflect.DeepEqual(queries["lang"], []string{"en"}) {
		t.Fatalf("expected Queries Values 'lang' to be ['en'], got %v", queries["lang"])
	}
	if len(queries["nonexistent"]) != 0 {
		t.Fatalf("expected non-existent query parameter in Queries to return empty slice, got %v", queries["nonexistent"])
	}
}

func TestRequest_Request(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.test/", nil)
	r := newRequest(nil, req)

	if r.Request() == nil {
		t.Fatalf("expected underlying http.Request non-nil")
	}
	if !reflect.DeepEqual(r.Request(), req) {
		t.Fatalf("expected Request() to return the original http.Request")
	}
}

func TestRequest_SetMaxBodyBytes(t *testing.T) {
	body := strings.NewReader(strings.Repeat("x", 100))
	req := httptest.NewRequest(http.MethodPost, "http://example.test/", body)

	r := newRequest(nil, req)
	r.SetMaxBodyBytes(50)

	rc, _ := r.Body()
	_, err := io.ReadAll(rc)
	if err == nil {
		t.Fatalf("expected MaxBytesError after SetMaxBodyBytes(50), got nil")
	}
}
