package web

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/url"

	"github.com/go-amwk/core"
)

// Request represents an HTTP request received by the application.
type Request struct {
	resource string

	app *Application
	req *http.Request
}

// newRequest creates a new Request instance from the application context and the underlying
// http.Request.
func newRequest(app *Application, r *http.Request) *Request {
	req := &Request{
		app: app,
		req: r,
	}

	return req
}

// Application returns the Application instance associated with this Request. This allows you to
// access the application context and its settings from within the request handling logic.
func (req *Request) Application() core.Application {
	if req.app == nil {
		return nil
	}
	return req.app
}

// Body returns the request body as a readable stream.
func (req *Request) Body() (io.ReadCloser, error) {
	return req.req.Body, nil
}

// ClientIP returns the client's IP address from the request.
func (req *Request) ClientIP() string {
	ip := req.req.RemoteAddr
	host, _, err := net.SplitHostPort(ip)
	if err != nil {
		return ip
	}

	return host
}

// ContentLength returns the length of the request body in bytes. If the length is unknown, it
// returns -1.
func (req *Request) ContentLength() int64 {
	return req.req.ContentLength
}

// Cookie returns the named cookie provided in the request.
func (req *Request) Cookie(name string) (*http.Cookie, error) {
	return req.req.Cookie(name)
}

// Cookies returns the cookies provided in the request.
func (req *Request) Cookies() []*http.Cookie {
	return req.req.Cookies()
}

// Context returns the context associated with the request.
func (req *Request) Context() context.Context {
	return req.req.Context()
}

// Header returns the value of the specified header field from the request. If the header is not
// present, it returns an empty string.
func (req *Request) Header(name string) string {
	return req.req.Header.Get(name)
}

// HeaderValues returns all values associated with the specified header field from the request. If
// the header is not present, it returns an empty slice.
func (req *Request) HeaderValues(name string) []string {
	return req.req.Header.Values(name)
}

// Headers returns all headers associated with the request.
func (req *Request) Headers() http.Header {
	// Make a copy of the headers to avoid modifying the original request headers
	headers := make(http.Header)
	for k, v := range req.req.Header {
		headers[k] = append(headers[k], v...)
	}

	return headers
}

// Method returns the HTTP method of the request (e.g., GET, POST, etc.).
func (req *Request) Method() string {
	return req.req.Method
}

// Protocol returns the HTTP protocol version used in the request (e.g., HTTP/1.1, HTTP/2, etc.).
func (req *Request) Protocol() string {
	return req.req.Proto
}

// Path returns the URL path of the request.
func (req *Request) Path() string {
	return req.req.URL.Path
}

// PathValue returns the value of the specified path parameter from the request. If the parameter
// is not present, it returns an empty string.
func (req *Request) PathValue(name string) string {
	return req.req.PathValue(name)
}

// SetPathValue sets the value of the specified path parameter in the request. This is typically
// used by the router to store path parameters extracted from the URL.
func (req *Request) SetPathValue(name, value string) {
	req.req.SetPathValue(name, value)
}

// Resource returns the resource associated with the request. The resource is a string that can be
// used to identify the type of request or the endpoint being accessed. It is typically set by the
// router based on the URL pattern matched for the request.
func (req *Request) Resource() string {
	return req.resource
}

// SetResource sets the resource associated with the request. This is typically used by the router
// to store the resource identifier based on the URL pattern matched for the request.
func (req *Request) SetResource(resource string) {
	req.resource = resource
}

// Query returns the value of the specified query parameter from the request URL. If the parameter
// is not present, it returns an empty string.
func (req *Request) Query(name string) string {
	return req.req.URL.Query().Get(name)
}

// QueryValues returns all values associated with the specified query parameter from the request
// URL. If the parameter is not present, it returns an empty slice.
func (req *Request) QueryValues(name string) []string {
	return req.req.URL.Query()[name]
}

// Queries returns all query parameters from the request URL as a url.Values map, where the keys
// are the parameter names and the values are slices of parameter values. If there are no query
// parameters, it returns an empty map.
func (req *Request) Queries() url.Values {
	// Make a copy of the query parameters to avoid modifying the original request URL query
	// parameters
	queries := make(url.Values)
	for key, values := range req.req.URL.Query() {
		queries[key] = append(queries[key], values...)
	}

	return queries
}

// Request returns the underlying http.Request associated with this Request. This can be used to
// access any additional information or functionality provided by the http.Request that is not
// exposed through the Request interface.
func (req *Request) Request() any {
	return req.req
}
