# HTTP Server from TCP

A custom HTTP/1.1 server implementation built using raw TCP sockets in Go, without using the standard library's `net/http` package.

## Features

- HTTP/1.1 request parsing (request line, headers, body)
- Concurrent connection handling
- Custom routing with handler functions
- Support for GET, POST, PUT, DELETE, OPTIONS methods
- Content-Length based body parsing
- Graceful shutdown handling

## Project Structure

```
├── cmd/main.go                # Server entry point
└── internal/
    ├── request/               # HTTP request parsing
    ├── response/              # HTTP response building
    ├── headers/               # Header parsing and handling
    └── server/                # TCP server implementation
```

## Running

```bash
go run cmd/main.go
```

Server runs on port 42069. Test with:
```bash
curl localhost:42069
curl localhost:42069/yourproblem  # Returns 400
curl localhost:42069/myproblem    # Returns 500
```

## Testing

```bash
go test ./...
```

## Technical Details

- Parses HTTP requests byte-by-byte from TCP streams
- Handles partial reads and buffering
- Validates HTTP/1.1 protocol compliance
- Implements proper CRLF line endings and header formatting

## Dependencies

- Standard library only (except `testify` for tests)
- No `net/http` usage
