package web

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestNewResponse_Headers(t *testing.T) {
	rr := httptest.NewRecorder()
	resp := newResponse(rr)

	resp.AddHeader("A", "1")
	resp.AddHeader("A", "2")
	if got := resp.GetHeader("A"); got != "1" {
		t.Fatalf("expected GetHeader '1', got %v", got)
	}

	resp.AddHeader("B", "old")
	resp.SetHeader("B", "v")
	if resp.GetHeader("B") != "v" {
		t.Fatalf("expected B=v, got %v", resp.GetHeader("B"))
	}

	resp.AddHeader("C", "v")
	resp.DelHeader("C")
	if resp.GetHeader("C") != "" {
		t.Fatalf("expected C deleted, got %v", resp.GetHeader("C"))
	}

	if h := resp.Headers(); !reflect.DeepEqual(h, http.Header{"A": []string{"1", "2"}, "B": []string{"v"}}) {
		t.Fatalf("unexpected Headers: %v", h)
	}

	resp.send()

	if rr.Header().Get("A") != "1" {
		t.Fatalf("expected sent header A=1, got %v", rr.Header().Get("A"))
	}
	if !reflect.DeepEqual(rr.Header()["A"], []string{"1", "2"}) {
		t.Fatalf("expected sent header A values ['1', '2'], got %v", rr.Header()["A"])
	}
	if rr.Header().Get("B") != "v" {
		t.Fatalf("expected sent header B=v, got %v", rr.Header().Get("B"))
	}
	if rr.Header().Get("C") != "" {
		t.Fatalf("expected sent header C deleted, got %v", rr.Header().Get("C"))
	}
}

func TestResponse_Write(t *testing.T) {
	rr := httptest.NewRecorder()
	resp := newResponse(rr)

	n, err := resp.Write([]byte("hello"))
	if err != nil || n != 5 {
		t.Fatalf("Write unexpected: n=%d err=%v", n, err)
	}

	if string(resp.body) != "hello" {
		t.Fatalf("expected body 'hello', got %q", string(resp.body))
	}

	n, err = resp.Write([]byte(" world"))
	if err != nil || n != 6 {
		t.Fatalf("Write unexpected: n=%d err=%v", n, err)
	}

	if string(resp.body) != "hello world" {
		t.Fatalf("expected body 'hello world', got %q", string(resp.body))
	}

	resp.send()

	if rr.Body.String() != "hello world" {
		t.Fatalf("expected sent body 'hello world', got %q", rr.Body.String())
	}
}

func TestResponse_Status(t *testing.T) {
	rr := httptest.NewRecorder()
	resp := newResponse(rr)

	resp.Status(201)
	if resp.StatusCode() != 201 {
		t.Fatalf("expected status 201, got %d", resp.StatusCode())
	}

	resp.send()

	if rr.Code != 201 {
		t.Fatalf("expected sent status 201, got %d", rr.Code)
	}
}

func TestResponse_Response(t *testing.T) {
	rr := httptest.NewRecorder()
	resp := newResponse(rr)
	if resp.Response() == nil {
		t.Fatalf("expected underlying response writer non-nil")
	}
	if !reflect.DeepEqual(resp.Response(), rr) {
		t.Fatalf("expected Response() to return the original response writer")
	}
}
