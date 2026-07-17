package web

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"

	"github.com/go-amwk/core"
)

// Request represents an HTTP request received by the application.
type Request struct {
	app *Application
	req *http.Request

	body         io.ReadCloser
	bodyOnce     sync.Once
	resource     string
	queries      url.Values
	queryOnce    sync.Once
	maxBodyBytes atomic.Int64
}

// newRequest creates a new Request instance from the application context and the underlying
// http.Request.
func newRequest(app *Application, r *http.Request) *Request {
	req := &Request{
		app: app,
		req: r,
	}
	if app != nil {
		req.maxBodyBytes.Store(app.MaxRequestBodyBytes())
	} else {
		req.maxBodyBytes.Store(MaxRequestBodyBytesDefault)
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
	req.bodyOnce.Do(func() {
		maxBytes := req.maxBodyBytes.Load()

		if maxBytes == MaxRequestBodyBytesUnlimited {
			req.body = req.req.Body
		} else {
			// Limit the request body size to the specified maximum bytes.
			req.body = http.MaxBytesReader(nil, req.req.Body, maxBytes)
		}
	})

	return req.body, nil
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
	queries := req.getQueries()
	return queries.Get(name)
}

// QueryValues returns all values associated with the specified query parameter from the request
// URL. If the parameter is not present, it returns an empty slice.
func (req *Request) QueryValues(name string) []string {
	queries := req.getQueries()
	if _, ok := queries[name]; !ok {
		return nil
	}

	values := make([]string, len(queries[name]))
	copy(values, queries[name])
	return values
}

// Queries returns all query parameters from the request URL as a url.Values map, where the keys
// are the parameter names and the values are slices of parameter values. If there are no query
// parameters, it returns an empty map.
func (req *Request) Queries() url.Values {
	queries := req.getQueries()
	// Make a copy of the query parameters to avoid modifying the original request URL query
	// parameters
	copiedQueries := make(url.Values)
	for key, values := range queries {
		copiedQueries[key] = append(copiedQueries[key], values...)
	}

	return copiedQueries
}

// Request returns the underlying http.Request associated with this Request. This can be used to
// access any additional information or functionality provided by the http.Request that is not
// exposed through the Request interface.
func (req *Request) Request() any {
	return req.req
}

// SetMaxBodyBytes sets the maximum body size in bytes for this request. A value of -1 indicates
// that there is no limit on the body size, while a value of 0 indicates that no body is allowed.
// This method should be called before reading the request body to ensure that the size limit is
// enforced correctly.
func (req *Request) SetMaxBodyBytes(size int64) {
	req.maxBodyBytes.Store(size)
}

// getQueries retrieves the query parameters from the request URL. It uses sync.Once to ensure that
// the query parameters are only parsed once, even if this method is called multiple times. The
// parsed query parameters are cached in the Request struct for future use.
func (req *Request) getQueries() url.Values {
	req.queryOnce.Do(func() {
		req.queries = req.req.URL.Query()
	})
	return req.queries
}
