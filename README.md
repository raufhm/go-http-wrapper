# Go HTTP Wrapper

A lightweight and flexible HTTP client wrapper for Go applications that simplifies making HTTP requests while providing additional features like middleware support, retry mechanisms, and timeout handling.

## Features

- Simple and intuitive API
- Middleware support for request/response modification
- Automatic retry mechanism
- Configurable timeout handling
- Context support
- Custom header management
- Response validation
- Error handling
- JSON request/response support

## Requirements

- Go 1.16 or higher

## Installation

```bash
go get github.com/raufhm/go-http-wrapper
```

## Usage

### Basic Request

```go
client := httpwrapper.NewClient()

// GET request
response, err := client.Get(context.Background(), "https://api.example.com/users")

// POST request with JSON body
payload := map[string]interface{}{
    "name": "John Doe",
    "email": "john@example.com",
}
response, err := client.Post(context.Background(), "https://api.example.com/users", payload)
```

### With Middleware

```go
client := httpwrapper.NewClient(
    httpwrapper.WithMiddleware(func(next http.RoundTripper) http.RoundTripper {
        return httpwrapper.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
            req.Header.Set("Authorization", "Bearer your-token")
            return next.RoundTrip(req)
        })
    }),
)
```

### With Retry

```go
client := httpwrapper.NewClient(
    httpwrapper.WithRetry(3, time.Second),
)
```

## Error Handling

The wrapper provides detailed error information:

```go
response, err := client.Get(context.Background(), "https://api.example.com/users")
if err != nil {
    if httpErr, ok := err.(*httpwrapper.Error); ok {
        fmt.Printf("Status: %d, Message: %s\n", httpErr.StatusCode, httpErr.Message)
    }
}
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
