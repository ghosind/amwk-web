package web

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/go-amwk/core"
	"github.com/go-amwk/engine"
)

const (
	// AddressDefault defines the default address for the application server to listen on. It is set
	// to ":8000", which means the server will listen on all available network interfaces on
	// port 8000.
	AddressDefault = ":8000"
	// IdleTimeoutDefault defines the default maximum amount of time to wait for the next request
	// when keep-alives are enabled. It is set to 60 seconds.
	IdleTimeoutDefault time.Duration = 60 * time.Second
	// MaxHeaderBytesDefault defines the default maximum size of request headers. It is set to
	// 1 MB.
	MaxHeaderBytesDefault int = 1 << 20
	// MaxResponseBodyBytesDefault defines the default maximum body size for responses. It is set
	// to 32 MB.
	MaxResponseBodyBytesDefault int64 = 1 << 25
	// MaxResponseBodyBytesUnlimited defines a special value indicating that there is no limit on
	// the body size for responses.
	MaxResponseBodyBytesUnlimited int64 = -1
	// ReadHeaderTimeoutDefault defines the default maximum duration for reading the headers of
	// the request. It is set to 10 seconds.
	ReadHeaderTimeoutDefault time.Duration = 10 * time.Second
	// ReadTimeoutDefault defines the default maximum duration for reading the entire request,
	// including the body. It is set to 30 seconds.
	ReadTimeoutDefault time.Duration = 30 * time.Second
	// ShutdownTimeoutDefault defines the default maximum duration for gracefully shutting down
	// the application server. It is set to 5 seconds.
	ShutdownTimeoutDefault time.Duration = 5 * time.Second
	// WriteTimeoutDefault defines the default maximum duration before timing out writes of the
	// response. It is set to 30 seconds.
	WriteTimeoutDefault time.Duration = 30 * time.Second
)

var (
	// shutdownListenSignalsDefault defines the default signals to listen for when shutting down the
	// application. It includes os.Interrupt, syscall.SIGTERM, and syscall.SIGQUIT.
	shutdownListenSignalsDefault = []os.Signal{os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT}
)

// Application is an HTTP web application to serve HTTP requests.
type Application struct {
	addr     string
	server   *http.Server
	handlers []core.HandlerFunc

	maxResponseBodyBytes  atomic.Int64
	enableShutdownSignal  atomic.Bool
	shutdownListenSignals []os.Signal
	shutdownTimeout       time.Duration
}

// defaultApp creates a new Application instance with default settings.
func defaultApp() *Application {
	app := &Application{}
	app.handlers = make([]core.HandlerFunc, 0)
	app.server = &http.Server{
		Handler:           app,
		ReadTimeout:       ReadTimeoutDefault,
		WriteTimeout:      WriteTimeoutDefault,
		IdleTimeout:       IdleTimeoutDefault,
		ReadHeaderTimeout: ReadHeaderTimeoutDefault,
		MaxHeaderBytes:    MaxHeaderBytesDefault,
	}
	app.
		SetAddress(AddressDefault).
		SetEnableShutdownSignal(true, shutdownListenSignalsDefault...).
		SetMaxResponseBodyBytes(MaxResponseBodyBytesDefault).
		SetShutdownTimeout(ShutdownTimeoutDefault)

	return app
}

// Default returns a default application instance with default settings.
func Default() *Application {
	app := defaultApp()

	return app
}

// New returns a new application instance with default settings. It allows for further
// customization before starting the server.
func New(opts ...Option) *Application {
	app := defaultApp()

	for _, opt := range opts {
		opt(app)
	}

	return app
}

// Start starts the application server and listens for incoming requests. It returns an error if
// it fails to start.
func (app *Application) Start() error {
	app.server.Addr = app.addr
	errCh := make(chan error, 1)
	go func() {
		errCh <- app.server.ListenAndServe()
	}()

	sig := app.listenShutdownSignals(app.shutdownListenSignals)
	defer func() {
		if sig != nil {
			signal.Stop(sig)
		}
	}()
	select {
	case <-sig:
		shutdownTimeout := app.shutdownTimeout
		if shutdownTimeout <= 0 {
			shutdownTimeout = ShutdownTimeoutDefault
		}

		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		return app.Shutdown(ctx)
	case err := <-errCh:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}

// Use adds the given handlers to the application.
func (app *Application) Use(handlers ...core.HandlerFunc) core.Application {
	app.handlers = append(app.handlers, handlers...)
	return app
}

// Close closes the application.
func (app *Application) Close() error {
	if app.server != nil {
		return app.server.Close()
	}
	return nil
}

// Shutdown gracefully shuts down the application server.
func (app *Application) Shutdown(ctx context.Context) error {
	if app.server != nil {
		return app.server.Shutdown(ctx)
	}
	return nil
}

// ServeHTTP implements the http.Handler interface to handle incoming HTTP requests.
func (app *Application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := newRequest(app, r)
	resp := newResponse(app, w)
	ctx := engine.NewContext(app, req, resp)
	ctx.Use(app.handlers...)

	ctx.Next()

	resp.send()
}

// Address returns the address the application is listening on.
func (app *Application) Address() string {
	return app.addr
}

// SetAddress sets the address for the application to listen on. Please call this method before
// starting the application server.
func (app *Application) SetAddress(addr string) *Application {
	app.addr = addr
	return app
}

// SetEnableShutdownSignal enables or disables the handling of shutdown signals for the
// application. If enabled, the application will listen for shutdown signals (e.g., os.Interrupt,
// syscall.SIGTERM, syscall.SIGQUIT) and gracefully shut down when such a signal is received.
// You can also specify custom signals to listen for by providing them as additional arguments.
func (app *Application) SetEnableShutdownSignal(enable bool, signals ...os.Signal) *Application {
	app.enableShutdownSignal.Store(enable)
	if len(signals) > 0 {
		app.shutdownListenSignals = signals
	}
	return app
}

// SetIdleTimeout sets the maximum amount of time to wait for the next request when keep-alives are
// enabled. It is recommended to call this method before starting the application server to avoid
// unexpected behavior.
func (app *Application) SetIdleTimeout(timeout time.Duration) *Application {
	if app.server != nil {
		app.server.IdleTimeout = timeout
	}
	return app
}

// SetMaxHeaderBytes sets the maximum size of request headers. It is recommended to call this
// method before starting the application server to avoid unexpected behavior.
func (app *Application) SetMaxHeaderBytes(size int) *Application {
	if app.server != nil {
		app.server.MaxHeaderBytes = size
	}
	return app
}

// MaxResponseBodyBytes returns the maximum body size for responses. If the size is set to
// MaxResponseBodyBytesUnlimited, there will be no limit on the body size.
func (app *Application) MaxResponseBodyBytes() int64 {
	return app.maxResponseBodyBytes.Load()
}

// SetMaxResponseBodyBytes sets the maximum body size for responses. If the size is set to
// MaxResponseBodyBytesUnlimited, there will be no limit on the body size.
// It would be better to call this method before starting the application server to avoid
// unexpected behavior.
func (app *Application) SetMaxResponseBodyBytes(size int64) *Application {
	app.maxResponseBodyBytes.Store(size)
	return app
}

// SetReadHeaderTimeout sets the maximum duration for reading the headers of the request. It is
// recommended to call this method before starting the application server to avoid unexpected
// behavior.
func (app *Application) SetReadHeaderTimeout(timeout time.Duration) *Application {
	if app.server != nil {
		app.server.ReadHeaderTimeout = timeout
	}
	return app
}

// SetReadTimeout sets the maximum duration for reading the entire request, including the body. It
// is recommended to call this method before starting the application server to avoid unexpected
// behavior.
func (app *Application) SetReadTimeout(timeout time.Duration) *Application {
	if app.server != nil {
		app.server.ReadTimeout = timeout
	}
	return app
}

// SetShutdownTimeout sets the maximum duration for gracefully shutting down the application
// server. It is recommended to call this method before starting the application server to avoid
// unexpected behavior.
func (app *Application) SetShutdownTimeout(timeout time.Duration) *Application {
	app.shutdownTimeout = timeout
	return app
}

// SetWriteTimeout sets the maximum duration before timing out writes of the response. It is
// recommended to call this method before starting the application server to avoid unexpected
// behavior.
func (app *Application) SetWriteTimeout(timeout time.Duration) *Application {
	if app.server != nil {
		app.server.WriteTimeout = timeout
	}
	return app
}

// listenShutdownSignals listens for shutdown signals if enabled and returns a channel that
// receives the signals. If shutdown signal handling is disabled, it returns nil.
func (app *Application) listenShutdownSignals(signals []os.Signal) chan os.Signal {
	if !app.enableShutdownSignal.Load() {
		return nil
	}

	if len(signals) == 0 {
		signals = shutdownListenSignalsDefault
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, signals...)
	return sig
}
