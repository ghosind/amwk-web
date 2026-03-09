package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/go-amwk/core"
)

func TestDefaultAndNew(t *testing.T) {
	d := Default()
	if d == nil {
		t.Fatalf("Default returned nil")
	}
	if _, ok := d.(*Application); !ok {
		t.Fatalf("Default did not return *Application")
	}

	n := New(WithAddress(":0"))
	if n == nil {
		t.Fatalf("New returned nil")
	}
	if _, ok := n.(*Application); !ok {
		t.Fatalf("New did not return *Application")
	}
}

func TestApplication_Start(t *testing.T) {
	app := New(WithAddress(":0")).(*Application)

	done := make(chan error, 1)
	go func() {
		done <- app.Start()
	}()

	time.Sleep(50 * time.Millisecond)

	if err := app.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown error: %v", err)
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Start returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("Start did not return in time")
	}
}

func TestApplication_Start_Signal(t *testing.T) {
	app := New(WithAddress(":0")).(*Application)

	done := make(chan error, 1)
	go func() {
		done <- app.Start()
	}()

	time.Sleep(50 * time.Millisecond)

	if err := syscall.Kill(os.Getpid(), syscall.SIGINT); err != nil {
		t.Fatalf("failed to send signal: %v", err)
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Start returned error after signal: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("Start did not return after signal in time")
	}
}

func TestApplication_Use(t *testing.T) {
	app := New().(*Application)

	ret := app.Use(func(c core.Context) error { return nil })
	if ret != app {
		t.Fatalf("Use did not return the application instance")
	}

	app.Use(func(c core.Context) error {
		c.SetHeader("X-Use", "ok")
		return nil
	})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://example.test/", nil)
	app.ServeHTTP(rr, req)
	if rr.Header().Get("X-Use") != "ok" {
		t.Fatalf("expected X-Use header 'ok', got %v", rr.Header().Get("X-Use"))
	}
}

func TestApplication_Close(t *testing.T) {
	app := New().(*Application)

	app.server = nil
	if err := app.Close(); err != nil {
		t.Fatalf("Close returned error on nil server: %v", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer ts.Close()
	app.server = ts.Config
	if err := app.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
}

func TestApplication_Shutdown(t *testing.T) {
	app := New().(*Application)

	app.server = nil
	if err := app.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown returned error on nil server: %v", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer ts.Close()
	app.server = ts.Config
	if err := app.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown returned error: %v", err)
	}
}

func TestApplication_ServeHTTP(t *testing.T) {
	app := New().(*Application)

	app.Use(func(c core.Context) error {
		c.SetHeader("X-Test", "v")
		c.Write([]byte("hello"))
		return nil
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://example.test/", nil)

	app.ServeHTTP(rr, req)

	if got := rr.Header().Get("X-Test"); got != "v" {
		t.Fatalf("expected X-Test header 'v', got %v", got)
	}
	if rr.Body.String() != "hello" {
		t.Fatalf("expected body 'hello', got %q", rr.Body.String())
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
}
