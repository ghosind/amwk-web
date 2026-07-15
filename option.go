package web

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
