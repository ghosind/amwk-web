package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/go-amwk/core"
)

func TestDefaultAndNew(t *testing.T) {
	d := Default()
	if d == nil {
		t.Fatalf("Default returned nil")
	}

	n := New(WithAddress(":0"))
	if n == nil {
		t.Fatalf("New returned nil")
	}
}

func TestApplication_Start(t *testing.T) {
	app := New(WithAddress(":0"))

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

func TestApplication_Use(t *testing.T) {
	app := New()

	ret := app.Use(func(c core.Context) error { return nil })
	if ret.(*Application) != app {
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
	app := New()

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
	app := New()

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
	app := New()

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

func TestApplication_Address(t *testing.T) {
	app := New(WithAddress(":8080"))
	if app.Address() != ":8080" {
		t.Fatalf("expected address ':8080', got %v", app.Address())
	}

	app.SetAddress(":9090")
	if app.Address() != ":9090" {
		t.Fatalf("expected address ':9090', got %v", app.Address())
	}
}

func TestApplication_EnableShutdownSignal(t *testing.T) {
	app := New()

	app.SetEnableShutdownSignal(true)
	if !app.enableShutdownSignal.Load() {
		t.Fatalf("expected enableShutdownSignal to be true")
	} else if !reflect.DeepEqual(app.shutdownListenSignals, shutdownListenSignalsDefault) {
		t.Fatalf("expected 3 default shutdown signals, got %d", len(app.shutdownListenSignals))
	}

	app.SetEnableShutdownSignal(true, os.Interrupt)
	if !app.enableShutdownSignal.Load() {
		t.Fatalf("expected enableShutdownSignal to be true")
	} else if !reflect.DeepEqual(app.shutdownListenSignals, []os.Signal{os.Interrupt}) {
		t.Fatalf("expected shutdown signals [os.Interrupt], got %v", app.shutdownListenSignals)
	}

	app.SetEnableShutdownSignal(false)
	if app.enableShutdownSignal.Load() {
		t.Fatalf("expected enableShutdownSignal to be false")
	}
}

func TestApplication_IdleTimeout(t *testing.T) {
	app := New()
	if app.server.IdleTimeout != IdleTimeoutDefault {
		t.Fatalf("expected default idle timeout %v, got %v", IdleTimeoutDefault, app.server.IdleTimeout)
	}

	app.SetIdleTimeout(120 * time.Second)
	if app.server.IdleTimeout != 120*time.Second {
		t.Fatalf("expected idle timeout 120s, got %v", app.server.IdleTimeout)
	}
}

func TestApplication_MaxHeaderBytes(t *testing.T) {
	app := New()
	if app.server.MaxHeaderBytes != MaxHeaderBytesDefault {
		t.Fatalf("expected default max header bytes %d, got %d", MaxHeaderBytesDefault, app.server.MaxHeaderBytes)
	}

	app.SetMaxHeaderBytes(4096)
	if app.server.MaxHeaderBytes != 4096 {
		t.Fatalf("expected max header bytes 4096, got %d", app.server.MaxHeaderBytes)
	}
}

func TestApplication_MaxRequestBodyBytes(t *testing.T) {
	app := Default()
	if app.MaxRequestBodyBytes() != MaxRequestBodyBytesDefault {
		t.Fatalf("expected default max request body bytes %d, got %d", MaxRequestBodyBytesDefault, app.MaxRequestBodyBytes())
	}

	app.SetMaxRequestBodyBytes(1024)
	if app.MaxRequestBodyBytes() != 1024 {
		t.Fatalf("expected max request body bytes 1024, got %d", app.MaxRequestBodyBytes())
	}

	app.SetMaxRequestBodyBytes(MaxRequestBodyBytesUnlimited)
	if app.MaxRequestBodyBytes() != MaxRequestBodyBytesUnlimited {
		t.Fatalf("expected max request body bytes unlimited, got %d", app.MaxRequestBodyBytes())
	}
}

func TestApplication_MaxResponseBodyBytes(t *testing.T) {
	app := Default()
	if app.MaxResponseBodyBytes() != MaxResponseBodyBytesDefault {
		t.Fatalf("expected default max body bytes %d, got %d", MaxResponseBodyBytesDefault, app.MaxResponseBodyBytes())
	}

	app.SetMaxResponseBodyBytes(1024)
	if app.MaxResponseBodyBytes() != 1024 {
		t.Fatalf("expected max body bytes 1024, got %d", app.MaxResponseBodyBytes())
	}

	app.SetMaxResponseBodyBytes(MaxResponseBodyBytesUnlimited)
	if app.MaxResponseBodyBytes() != MaxResponseBodyBytesUnlimited {
		t.Fatalf("expected max body bytes unlimited, got %d", app.MaxResponseBodyBytes())
	}
}

func TestApplication_ReadHeaderTimeout(t *testing.T) {
	app := New()
	if app.server.ReadHeaderTimeout != ReadHeaderTimeoutDefault {
		t.Fatalf("expected default read header timeout %v, got %v", ReadHeaderTimeoutDefault, app.server.ReadHeaderTimeout)
	}

	app.SetReadHeaderTimeout(5 * time.Second)
	if app.server.ReadHeaderTimeout != 5*time.Second {
		t.Fatalf("expected read header timeout 5s, got %v", app.server.ReadHeaderTimeout)
	}
}

func TestApplication_ReadTimeout(t *testing.T) {
	app := New()
	if app.server.ReadTimeout != ReadTimeoutDefault {
		t.Fatalf("expected default read timeout %v, got %v", ReadTimeoutDefault, app.server.ReadTimeout)
	}

	app.SetReadTimeout(60 * time.Second)
	if app.server.ReadTimeout != 60*time.Second {
		t.Fatalf("expected read timeout 60s, got %v", app.server.ReadTimeout)
	}
}

func TestApplication_ShutdownTimeout(t *testing.T) {
	app := New()
	if app.shutdownTimeout != ShutdownTimeoutDefault {
		t.Fatalf("expected default shutdown timeout %v, got %v", ShutdownTimeoutDefault, app.shutdownTimeout)
	}

	app.SetShutdownTimeout(10 * time.Second)
	if app.shutdownTimeout != 10*time.Second {
		t.Fatalf("expected shutdown timeout 10s, got %v", app.shutdownTimeout)
	}

	// zero/negative values are allowed at setter level; Start() will fall back to default.
	app.SetShutdownTimeout(0)
	if app.shutdownTimeout != 0 {
		t.Fatalf("expected shutdown timeout 0, got %v", app.shutdownTimeout)
	}
	app.SetShutdownTimeout(-1 * time.Second)
	if app.shutdownTimeout != -1*time.Second {
		t.Fatalf("expected shutdown timeout -1s, got %v", app.shutdownTimeout)
	}
}

func TestApplication_WriteTimeout(t *testing.T) {
	app := New()
	if app.server.WriteTimeout != WriteTimeoutDefault {
		t.Fatalf("expected default write timeout %v, got %v", WriteTimeoutDefault, app.server.WriteTimeout)
	}

	app.SetWriteTimeout(60 * time.Second)
	if app.server.WriteTimeout != 60*time.Second {
		t.Fatalf("expected write timeout 60s, got %v", app.server.WriteTimeout)
	}
}
