package web

import (
	"bytes"
	"io"
	"net/http"
	"sync/atomic"

	"github.com/go-amwk/core"
)

type Response struct {
	app          *Application
	rw           http.ResponseWriter
	statusCode   int
	headers      http.Header
	body         *bytes.Buffer
	maxBodyBytes atomic.Int64
}

// newResponse creates a new Response instance with application context and http.ResponseWriter.
// It initializes the headers and body buffer for the response.
func newResponse(app *Application, rw http.ResponseWriter) *Response {
	resp := &Response{
		app:     app,
		rw:      rw,
		headers: make(http.Header),
		body:    bytes.NewBuffer(nil),
	}

	if app != nil {
		resp.maxBodyBytes.Store(app.MaxResponseBodyBytes())
	} else {
		resp.maxBodyBytes.Store(MaxResponseBodyBytesDefault)
	}

	return resp
}

// Application returns the Application instance associated with this Response. This allows you to
// access the application context and its settings from within the response handling logic.
func (resp *Response) Application() core.Application {
	if resp.app == nil {
		return nil
	}
	return resp.app
}

// AddHeader adds a header field with the specified key and value to the response. If the header
// field already exists, the new value will be appended to the existing values for that header.
func (resp *Response) AddHeader(key, value string) {
	resp.headers.Add(key, value)
}

// SetHeader sets a header field with the specified key and value in the response. If the header
// field already exists, its value will be replaced with the new value.
func (resp *Response) SetHeader(key, value string) {
	resp.headers.Set(key, value)
}

// GetHeader returns the value of the specified header field from the response. If the header field
// is not present, it returns an empty string.
func (resp *Response) GetHeader(key string) string {
	return resp.headers.Get(key)
}

// DelHeader deletes the specified header field from the response. If the header field is not
// present, this method does nothing.
func (resp *Response) DelHeader(key string) {
	resp.headers.Del(key)
}

// Headers returns the http.Header map containing all the header fields and their values in the
// response. This allows you to access and manipulate the headers of the response as needed before
// sending it back to the client.
func (resp *Response) Headers() http.Header {
	return resp.headers
}

// Write writes the given data to the response body. It appends the data to any existing content in
// the body. It returns the number of bytes written and any error encountered during the write
// operation.
func (resp *Response) Write(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}

	maxBytes := resp.maxBodyBytes.Load()

	if maxBytes == MaxResponseBodyBytesUnlimited {
		return resp.body.Write(data)
	}

	if int64(resp.body.Len())+int64(len(data)) > maxBytes {
		return 0, ErrResponseTooLarge
	}
	return resp.body.Write(data)
}

// Status sets the HTTP status code for the response. This method allows you to specify the status
// code that should be sent back to the client along with the response. If you do not set a status
// code, the default status code of 200 OK will be used when the response is sent.
func (resp *Response) Status(code int) {
	resp.statusCode = code
}

// StatusCode returns the HTTP status code that has been set for the response. If no status code
// has been set, it returns 0, which indicates that the default status code of 200 OK will be used
// when the response is sent.
func (resp *Response) StatusCode() int {
	return resp.statusCode
}

// Response returns the underlying http.ResponseWriter associated with this Response. This can be
// used to access any additional information or functionality provided by the http.ResponseWriter
// that is not exposed through the Response interface.
//
// Note: Do not manipulate the returned http.ResponseWriter directly, as it may interfere with the
// proper handling of the response by the framework. Use the methods provided by the Response
// interface to manipulate the response instead.
func (resp *Response) Response() any {
	return resp.rw
}

// SetMaxBodyBytes sets the maximum body size in bytes for this response. A value of -1 indicates
// that there is no limit on the body size, while a value of 0 indicates that no body is allowed.
func (resp *Response) SetMaxBodyBytes(size int64) {
	resp.maxBodyBytes.Store(size)
}

// send sends the response to the client by writing the headers, status code, and body to the
// http.ResponseWriter. It returns an error if there is an issue while writing the response.
func (resp *Response) send() error {
	for key, values := range resp.headers {
		for _, value := range values {
			resp.rw.Header().Add(key, value)
		}
	}

	if resp.statusCode != 0 {
		resp.rw.WriteHeader(resp.statusCode)
	} else {
		resp.rw.WriteHeader(http.StatusOK)
	}

	if resp.body.Len() > 0 {
		_, err := io.Copy(resp.rw, resp.body)
		if err != nil {
			return err
		}
		resp.body.Reset()
	}

	return nil
}
