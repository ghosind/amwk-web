package web

import (
	"os"
	"time"
)

// Option is a function that configures the Application.
type Option func(*Application)

// WithAddress sets the address for the application server to listen on.
func WithAddress(addr string) Option {
	return func(app *Application) {
		app.SetAddress(addr)
	}
}

// WithEnableShutdownSignal configures whether the application should listen for shutdown signals.
// If enabled, the application will listen for the specified signals (e.g., os.Interrupt,
// syscall.SIGTERM). By default, the application listens for os.Interrupt, syscall.SIGTERM, and
// syscall.SIGQUIT.
func WithEnableShutdownSignal(enable bool, signals ...os.Signal) Option {
	return func(app *Application) {
		app.SetEnableShutdownSignal(enable, signals...)
	}
}

// WithIdleTimeout sets the maximum amount of time to wait for the next request when keep-alives
// are enabled. It is recommended to call this method before starting the application server to
// avoid unexpected behavior.
func WithIdleTimeout(timeout time.Duration) Option {
	return func(app *Application) {
		app.SetIdleTimeout(timeout)
	}
}

// WithMaxHeaderBytes sets the maximum size of request headers. If the size is set to
// MaxHeaderBytesUnlimited, there will be no limit on the header size. It is recommended to call
// this method before starting the application server to avoid unexpected behavior.
func WithMaxHeaderBytes(size int) Option {
	return func(app *Application) {
		app.SetMaxHeaderBytes(size)
	}
}

// WithMaxResponseBodyBytes sets the maximum body size for responses. If the size is set to
// MaxResponseBodyBytesUnlimited, there will be no limit on the body size.
func WithMaxResponseBodyBytes(size int64) Option {
	return func(app *Application) {
		app.SetMaxResponseBodyBytes(size)
	}
}

// WithReadHeaderTimeout sets the maximum duration for reading the headers of the request. It is
// recommended to call this method before starting the application server to avoid unexpected
// behavior.
func WithReadHeaderTimeout(timeout time.Duration) Option {
	return func(app *Application) {
		app.SetReadHeaderTimeout(timeout)
	}
}

// WithReadTimeout sets the maximum duration for reading the entire request, including the body. It
// is recommended to call this method before starting the application server to avoid unexpected
// behavior.
func WithReadTimeout(timeout time.Duration) Option {
	return func(app *Application) {
		app.SetReadTimeout(timeout)
	}
}

// WithShutdownTimeout sets the maximum duration for gracefully shutting down the application
// server. It is recommended to call this method before starting the application server to avoid
// unexpected behavior.
func WithShutdownTimeout(timeout time.Duration) Option {
	return func(app *Application) {
		app.SetShutdownTimeout(timeout)
	}
}

// WithWriteTimeout sets the maximum duration before timing out writes of the response. It is
// recommended to call this method before starting the application server to avoid unexpected
// behavior.
func WithWriteTimeout(timeout time.Duration) Option {
	return func(app *Application) {
		app.SetWriteTimeout(timeout)
	}
}
