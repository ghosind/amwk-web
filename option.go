package web

// Option is a function that configures the Application.
type Option func(*Application)

// WithAddress sets the address for the application server to listen on.
func WithAddress(addr string) Option {
	return func(app *Application) {
		app.addr = addr
	}
}
