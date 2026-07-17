# HTTP Adapter for AMWK Framework

`amwk/web` is the HTTP adapter for the AMWK framework, providing the net adapter for building web applications.

## Installation

Please ensure you have installed Go 1.22 or the later version, and use the following command to install the amwk framework and the web adapter:

```bash
go get github.com/go-amwk/web
go mod tidy
```

## Getting Started

Here is a simple example demonstrating how to use the `web` package to create a basic web server.

```go
package main

import (
  "github.com/go-amwk/core"
  "github.com/go-amwk/web"
)

func main() {
  // Create a new web application instance.
  app := web.Default()

  // Add a simple handler that responds with "Hello, World!" to any request.
  app.Use(func(ctx core.Context) error {
    ctx.Write([]byte("Hello, World!"))
    return nil
  })

  // Start the server
  app.Start()
}
```

The server will listen on the default port (8000) and respond with "Hello, World!" to any incoming HTTP request.

For more detailed examples and usage, please refer to the [examples](https://github.com/go-amwk/examples) repository.

## Configuration

The `web` package can be configured using various options when creating a new application instance. Here are some of the available configuration options:

- `WithAddress(addr string)`: Sets the address for the server to listen on. Default is `:8000`.
- `WithEnableShutdownSignal(enable bool, signals ...os.Signal)`: Enables or disables shutdown signal handling. By default, it is enabled and listens for `os.Interrupt`, `syscall.SIGTERM`, and `syscall.SIGQUIT` signals.
- `WithIdleTimeout(timeout time.Duration)`: Sets the maximum amount of time to wait for the next request when keep-alives are enabled. Default is `60s`.
- `WithMaxHeaderBytes(size int)`: Sets the maximum size of request headers in bytes. Default is `1 MB` (1048576 bytes).
- `WithMaxResponseBodyBytes(size int64)`: Sets the maximum size of the response body in bytes. Default is `32 MB`. Set to `-1` for unlimited body size.
- `WithReadHeaderTimeout(timeout time.Duration)`: Sets the maximum duration for reading request headers. Default is `10s`.
- `WithReadTimeout(timeout time.Duration)`: Sets the maximum duration for reading the entire request, including the body. Default is `30s`.
- `WithShutdownTimeout(timeout time.Duration)`: Sets the maximum duration for gracefully shutting down the server. Default is `5s`.
- `WithWriteTimeout(timeout time.Duration)`: Sets the maximum duration before timing out writes of the response. Default is `30s`.

## Contributing

- Contributions are welcome. Please open issues for bugs or feature requests.
- For code changes, fork the repository, create a topic branch, and submit a pull request with a clear description of the change.
- Follow the repository's code style and include tests where appropriate.
- Run `go vet` and `go test ./...` before submitting a PR.

## License

The project is licensed under the Apache 2.0 License. See the [LICENSE](LICENSE) file for details.
