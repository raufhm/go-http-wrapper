# Go HTTP Wrapper

A lightweight and flexible HTTP client wrapper for Go applications with built-in retry mechanism and New Relic instrumentation.

## Features

- Built-in exponential backoff retry mechanism
- New Relic integration
- Configurable timeouts
- Query parameters support
- JSON request body handling
- Custom headers support
- Context support
- HTTP methods: GET, POST, PUT, PATCH, DELETE

## Requirements

- Go 1.16 or higher
- New Relic Go Agent v3
- Echo v4 (for MIME types)

## Installation

```bash
go get github.com/raufhm/go-http-wrapper
```

## Usage

### Basic Client Setup

```go
// Create a new client with base URL
client := httpwrapper.New(
    "https://api.example.com",
    WithTimeout(30 * time.Second),
    WithHeaders(map[string]string{
        "Authorization": "Bearer token",
    }),
)
```

### Making Requests

```go
// GET request with query parameters
params := map[string][]string{
    "page": {"1"},
    "limit": {"10"},
}
resp, err := client.Get(ctx, "/users", WithQueryParams(params))

// POST request with JSON body
body := map[string]interface{}{
    "name": "John Doe",
    "email": "john@example.com",
}
resp, err := client.Post(ctx, "/users", WithBodyRequest(body))
```

### Custom Backoff Configuration

```go
expBackoff := backoff.NewExponentialBackOff()
expBackoff.MaxElapsedTime = 1 * time.Minute

client := httpwrapper.New(
    "https://api.example.com",
    WithBackoff(expBackoff),
)
```

### Error Handling

```go
resp, err := client.Get(ctx, "/users")
if err != nil {
    // Errors include:
    // - Request creation errors
    // - Network errors
    // - Non-2xx responses (with status code)
    // - Context cancellation
    return err
}
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
