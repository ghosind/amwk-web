package web

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-amwk/core"
	"github.com/go-amwk/engine"
)

const (
	// MaxResponseBodyBytesDefault defines the default maximum body size for responses. It is set
	// to 32 MB.
	MaxResponseBodyBytesDefault int64 = 32 * 1024 * 1024
	// MaxResponseBodyBytesUnlimited defines a special value indicating that there is no limit on
	// the body size for responses.
	MaxResponseBodyBytesUnlimited int64 = -1
)

// Application is an HTTP web application to serve HTTP requests.
type Application struct {
	addr     string
	server   *http.Server
	handlers []core.HandlerFunc

	maxResponseBodyBytes *int64
}

// defaultApp creates a new Application instance with default settings.
func defaultApp() *Application {
	app := &Application{}
	app.handlers = make([]core.HandlerFunc, 0)
	app.addr = ":8000"
	app.server = &http.Server{
		Handler: app,
	}

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

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer signal.Stop(sig)
	select {
	case <-sig:
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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
func (app *Application) SetAddress(addr string) {
	app.addr = addr
}

// MaxResponseBodyBytes returns the maximum body size for responses. If the size is set to
// MaxResponseBodyBytesUnlimited, there will be no limit on the body size.
func (app *Application) MaxResponseBodyBytes() int64 {
	if app.maxResponseBodyBytes == nil {
		return MaxResponseBodyBytesDefault
	}
	return *app.maxResponseBodyBytes
}

// SetMaxResponseBodyBytes sets the maximum body size for responses. If the size is set to
// MaxResponseBodyBytesUnlimited, there will be no limit on the body size.
// It would be better to call this method before starting the application server to avoid
// unexpected behavior.
func (app *Application) SetMaxResponseBodyBytes(size int64) {
	if app.maxResponseBodyBytes == nil {
		app.maxResponseBodyBytes = new(int64)
	}
	*app.maxResponseBodyBytes = size
}
