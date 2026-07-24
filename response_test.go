package web

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestResponse_Application(t *testing.T) {
	app := Default()
	resp := newResponse(app, httptest.NewRecorder())
	if resp.Application() != app {
		t.Fatalf("expected Application() to return the original application instance")
	}

	resp = newResponse(nil, httptest.NewRecorder())
	if resp.Application() != nil {
		t.Fatalf("expected Application() to return nil when no application is associated")
	}
}

func TestResponse_Headers(t *testing.T) {
	rr := httptest.NewRecorder()
	resp := newResponse(nil, rr)

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
	resp := newResponse(nil, rr)

	n, err := resp.Write([]byte{}) // empty
	if err != nil || n != 0 {
		t.Fatalf("Write unexpected: n=%d err=%v", n, err)
	}

	n, err = resp.Write(nil) // nil
	if err != nil || n != 0 {
		t.Fatalf("Write unexpected: n=%d err=%v", n, err)
	}

	n, err = resp.Write([]byte("hello"))
	if err != nil || n != 5 {
		t.Fatalf("Write unexpected: n=%d err=%v", n, err)
	}

	if resp.body.String() != "hello" {
		t.Fatalf("expected body 'hello', got %q", resp.body.String())
	}

	n, err = resp.Write([]byte(" world"))
	if err != nil || n != 6 {
		t.Fatalf("Write unexpected: n=%d err=%v", n, err)
	}

	if resp.body.String() != "hello world" {
		t.Fatalf("expected body 'hello world', got %q", resp.body.String())
	}

	resp.send()

	if rr.Body.String() != "hello world" {
		t.Fatalf("expected sent body 'hello world', got %q", rr.Body.String())
	}
}

func TestResponse_Write_LargeBody(t *testing.T) {
	app := New(WithMaxResponseBodyBytes(11)) // Set max body size to 11 bytes
	rr := httptest.NewRecorder()
	resp := newResponse(app, rr)

	n, err := resp.Write([]byte("hello"))
	if err != nil || n != 5 {
		t.Fatalf("Write unexpected: n=%d err=%v", n, err)
	}

	n, err = resp.Write([]byte(" "))
	if err != nil || n != 1 {
		t.Fatalf("Write unexpected: n=%d err=%v", n, err)
	}

	n, err = resp.Write([]byte("world!"))
	if err == nil || !errors.Is(err, ErrResponseTooLarge) || n != 0 {
		t.Fatalf("expected error ErrResponseTooLarge when writing large body, got n=%d err=%v", n, err)
	}

	n, err = resp.Write([]byte("world"))
	if err != nil || n != 5 {
		t.Fatalf("Write unexpected: n=%d err=%v", n, err)
	}

	n, err = resp.Write([]byte("!"))
	if err == nil || !errors.Is(err, ErrResponseTooLarge) || n != 0 {
		t.Fatalf("expected error ErrResponseTooLarge when writing large body, got n=%d err=%v", n, err)
	}

	if resp.body.String() != "hello world" {
		t.Fatalf("expected body 'hello world', got %q", resp.body.String())
	}

	resp.send()

	if rr.Body.String() != "hello world" {
		t.Fatalf("expected sent body 'hello world', got %q", rr.Body.String())
	}
}

func TestResponse_Write_Unlimited(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	app := New(WithMaxResponseBodyBytes(MaxResponseBodyBytesUnlimited)) // Set max body size to unlimited
	rr := httptest.NewRecorder()
	resp := newResponse(app, rr)

	largeData := make([]byte, MaxResponseBodyBytesDefault+1) // Create data larger than default max size

	n, err := resp.Write(largeData)
	if err != nil || n != len(largeData) {
		t.Fatalf("Write unexpected: n=%d err=%v", n, err)
	}
}

func TestResponse_Status(t *testing.T) {
	rr := httptest.NewRecorder()
	resp := newResponse(nil, rr)

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
	resp := newResponse(nil, rr)
	if resp.Response() == nil {
		t.Fatalf("expected underlying response writer non-nil")
	}
	if !reflect.DeepEqual(resp.Response(), rr) {
		t.Fatalf("expected Response() to return the original response writer")
	}
}

func TestResponse_SetMaxBodyBytes(t *testing.T) {
	rr := httptest.NewRecorder()
	resp := newResponse(nil, rr)

	// default is MaxResponseBodyBytesDefault
	resp.SetMaxBodyBytes(10)

	n, err := resp.Write([]byte("hello"))
	if err != nil || n != 5 {
		t.Fatalf("Write unexpected: n=%d err=%v", n, err)
	}

	n, err = resp.Write([]byte(" world"))
	if err == nil || !errors.Is(err, ErrResponseTooLarge) || n != 0 {
		t.Fatalf("expected ErrResponseTooLarge after SetMaxBodyBytes(10), got n=%d err=%v", n, err)
	}

	// set to unlimited, should allow larger writes
	resp.SetMaxBodyBytes(MaxResponseBodyBytesUnlimited)
	n, err = resp.Write([]byte(" world"))
	if err != nil || n != 6 {
		t.Fatalf("Write unexpected after unlimited: n=%d err=%v", n, err)
	}

	// set to 0 should block all writes
	newResp := newResponse(nil, httptest.NewRecorder())
	newResp.SetMaxBodyBytes(0)
	n, err = newResp.Write([]byte("x"))
	if err == nil || !errors.Is(err, ErrResponseTooLarge) || n != 0 {
		t.Fatalf("expected ErrResponseTooLarge when maxBodyBytes=0, got n=%d err=%v", n, err)
	}
}

func TestResponse_WriteString(t *testing.T) {
	rr := httptest.NewRecorder()
	resp := newResponse(nil, rr)

	n, err := resp.WriteString("")
	if err != nil || n != 0 {
		t.Fatalf("WriteString empty: n=%d err=%v", n, err)
	}

	n, err = resp.WriteString("hello")
	if err != nil || n != 5 {
		t.Fatalf("WriteString: n=%d err=%v", n, err)
	}

	if resp.body.String() != "hello" {
		t.Fatalf("expected body 'hello', got %q", resp.body.String())
	}

	n, err = resp.WriteString(" world")
	if err != nil || n != 6 {
		t.Fatalf("WriteString: n=%d err=%v", n, err)
	}

	if resp.body.String() != "hello world" {
		t.Fatalf("expected body 'hello world', got %q", resp.body.String())
	}

	resp.send()

	if rr.Body.String() != "hello world" {
		t.Fatalf("expected sent body 'hello world', got %q", rr.Body.String())
	}
}

func TestResponse_WriteString_LargeBody(t *testing.T) {
	app := New(WithMaxResponseBodyBytes(11))
	rr := httptest.NewRecorder()
	resp := newResponse(app, rr)

	n, err := resp.WriteString("hello")
	if err != nil || n != 5 {
		t.Fatalf("WriteString: n=%d err=%v", n, err)
	}

	n, err = resp.WriteString(" world!")
	if err == nil || !errors.Is(err, ErrResponseTooLarge) || n != 0 {
		t.Fatalf("expected ErrResponseTooLarge, got n=%d err=%v", n, err)
	}

	// body should remain unchanged after failed write
	if resp.body.String() != "hello" {
		t.Fatalf("expected body 'hello' after failed write, got %q", resp.body.String())
	}

	resp.send()

	if rr.Body.String() != "hello" {
		t.Fatalf("expected sent body 'hello', got %q", rr.Body.String())
	}
}

func TestResponse_WriteString_Unlimited(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	app := New(WithMaxResponseBodyBytes(MaxResponseBodyBytesUnlimited))
	rr := httptest.NewRecorder()
	resp := newResponse(app, rr)

	largeStr := string(make([]byte, MaxResponseBodyBytesDefault+1))

	n, err := resp.WriteString(largeStr)
	if err != nil || n != len(largeStr) {
		t.Fatalf("WriteString unlimited: n=%d err=%v", n, err)
	}
}

func TestResponse_Size(t *testing.T) {
	rr := httptest.NewRecorder()
	resp := newResponse(nil, rr)

	if resp.Size() != 0 {
		t.Fatalf("expected Size 0, got %d", resp.Size())
	}

	resp.Write([]byte("hello"))
	if resp.Size() != 5 {
		t.Fatalf("expected Size 5, got %d", resp.Size())
	}

	resp.WriteString(" world")
	if resp.Size() != 11 {
		t.Fatalf("expected Size 11, got %d", resp.Size())
	}
}

func TestResponse_Written(t *testing.T) {
	rr := httptest.NewRecorder()
	resp := newResponse(nil, rr)

	if len(resp.Written()) != 0 {
		t.Fatalf("expected empty Written, got %v", resp.Written())
	}

	resp.Write([]byte("hello"))
	if string(resp.Written()) != "hello" {
		t.Fatalf("expected Written 'hello', got %q", string(resp.Written()))
	}

	// verify that mutating the returned slice does not affect the body
	data := resp.Written()
	data[0] = 'x'
	if string(resp.Written()) != "hello" {
		t.Fatalf("Written should return a copy; mutation must not affect body, got %q", string(resp.Written()))
	}

	resp.WriteString(" world")
	if string(resp.Written()) != "hello world" {
		t.Fatalf("expected Written 'hello world', got %q", string(resp.Written()))
	}
}
