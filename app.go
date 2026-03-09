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

// Application is an HTTP web application to serve HTTP requests.
type Application struct {
	addr     string
	server   *http.Server
	handlers []core.HandlerFunc
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
func Default() core.Application {
	app := defaultApp()

	return app
}

// New returns a new application instance with default settings. It allows for further
// customization before starting the server.
func New(opts ...Option) core.Application {
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
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
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
	req := newRequest(r)
	resp := newResponse(w)
	ctx := engine.NewContext(app, req, resp)
	ctx.Use(app.handlers...)

	ctx.Next()

	resp.send()
}
