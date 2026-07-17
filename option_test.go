package web

import (
	"os"
	"testing"
	"time"
)

func TestWithAddress(t *testing.T) {
	app := defaultApp()
	opt := WithAddress(":1234")
	opt(app)
	if app.Address() != ":1234" {
		t.Fatalf("expected addr :1234, got %v", app.Address())
	}
}

func TestWithEnableShutdownSignal(t *testing.T) {
	app := defaultApp()

	// disable shutdown signal
	opt := WithEnableShutdownSignal(false)
	opt(app)
	if app.enableShutdownSignal.Load() {
		t.Fatalf("expected shutdown signal disabled")
	}

	// enable with custom signal
	opt = WithEnableShutdownSignal(true, os.Interrupt)
	opt(app)
	if !app.enableShutdownSignal.Load() {
		t.Fatalf("expected shutdown signal enabled")
	}
	if len(app.shutdownListenSignals) != 1 || app.shutdownListenSignals[0] != os.Interrupt {
		t.Fatalf("expected shutdown listen signals to contain os.Interrupt")
	}
}

func TestWithIdleTimeout(t *testing.T) {
	app := defaultApp()
	opt := WithIdleTimeout(120 * time.Second)
	opt(app)
	if app.server.IdleTimeout != 120*time.Second {
		t.Fatalf("expected idle timeout 120s, got %v", app.server.IdleTimeout)
	}
}

func TestWithMaxHeaderBytes(t *testing.T) {
	app := defaultApp()
	opt := WithMaxHeaderBytes(4096)
	opt(app)
	if app.server.MaxHeaderBytes != 4096 {
		t.Fatalf("expected max header bytes 4096, got %d", app.server.MaxHeaderBytes)
	}
}

func TestWithMaxRequestBodyBytes(t *testing.T) {
	app := defaultApp()
	opt := WithMaxRequestBodyBytes(1024)
	opt(app)
	if app.MaxRequestBodyBytes() != 1024 {
		t.Fatalf("expected max request body bytes 1024, got %d", app.MaxRequestBodyBytes())
	}

	opt = WithMaxRequestBodyBytes(MaxRequestBodyBytesUnlimited)
	opt(app)
	if app.MaxRequestBodyBytes() != MaxRequestBodyBytesUnlimited {
		t.Fatalf("expected max request body bytes unlimited, got %d", app.MaxRequestBodyBytes())
	}
}

func TestWithMaxResponseBodyBytes(t *testing.T) {
	app := defaultApp()
	opt := WithMaxResponseBodyBytes(1024)
	opt(app)
	if app.MaxResponseBodyBytes() != 1024 {
		t.Fatalf("expected max body bytes 1024, got %d", app.MaxResponseBodyBytes())
	}

	opt = WithMaxResponseBodyBytes(MaxResponseBodyBytesUnlimited)
	opt(app)
	if app.MaxResponseBodyBytes() != MaxResponseBodyBytesUnlimited {
		t.Fatalf("expected max body bytes unlimited, got %d", app.MaxResponseBodyBytes())
	}
}

func TestWithReadHeaderTimeout(t *testing.T) {
	app := defaultApp()
	opt := WithReadHeaderTimeout(5 * time.Second)
	opt(app)
	if app.server.ReadHeaderTimeout != 5*time.Second {
		t.Fatalf("expected read header timeout 5s, got %v", app.server.ReadHeaderTimeout)
	}
}

func TestWithReadTimeout(t *testing.T) {
	app := defaultApp()
	opt := WithReadTimeout(60 * time.Second)
	opt(app)
	if app.server.ReadTimeout != 60*time.Second {
		t.Fatalf("expected read timeout 60s, got %v", app.server.ReadTimeout)
	}
}

func TestWithShutdownTimeout(t *testing.T) {
	app := defaultApp()
	opt := WithShutdownTimeout(10 * time.Second)
	opt(app)
	if app.shutdownTimeout != 10*time.Second {
		t.Fatalf("expected shutdown timeout 10s, got %v", app.shutdownTimeout)
	}

	// zero/negative values are allowed at setter level; Start() will fall back to default.
	opt = WithShutdownTimeout(0)
	opt(app)
	if app.shutdownTimeout != 0 {
		t.Fatalf("expected shutdown timeout 0, got %v", app.shutdownTimeout)
	}
	opt = WithShutdownTimeout(-1 * time.Second)
	opt(app)
	if app.shutdownTimeout != -1*time.Second {
		t.Fatalf("expected shutdown timeout -1s, got %v", app.shutdownTimeout)
	}
}

func TestWithWriteTimeout(t *testing.T) {
	app := defaultApp()
	opt := WithWriteTimeout(60 * time.Second)
	opt(app)
	if app.server.WriteTimeout != 60*time.Second {
		t.Fatalf("expected write timeout 60s, got %v", app.server.WriteTimeout)
	}
}
