package web

import "os"

// Option is a function that configures the Application.
type Option func(*Application)

// WithAddress sets the address for the application server to listen on.
func WithAddress(addr string) Option {
	return func(app *Application) {
		app.SetAddress(addr)
	}
}

// WithMaxResponseBodyBytes sets the maximum body size for responses. If the size is set to
// MaxResponseBodyBytesUnlimited, there will be no limit on the body size.
func WithMaxResponseBodyBytes(size int64) Option {
	return func(app *Application) {
		app.SetMaxResponseBodyBytes(size)
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
