package response

import (
	"fmt"
	"io"

	"github.com/yus-works/tcp-to-http/internal/headers"
	"github.com/yus-works/tcp-to-http/internal/request"
)

type StatusCode int
const (
	StatusOK StatusCode = 200
	StatusBadRequest StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

var statusText = map[StatusCode]string{
    StatusOK:                  "OK",
    StatusBadRequest:          "Bad Request",
    StatusInternalServerError: "Internal Server Error",
}

func (s StatusCode) String() string {
    if text, ok := statusText[s]; ok {
        return text
    }
    return ""
}

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	_, err := fmt.Fprintf(w,
		"HTTP/1.1 %d %s\r\n",
		statusCode, statusCode.String(),
	)
	if err != nil {
		return fmt.Errorf("Failed to write status line: %w", err)
	}
	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()

	headers.Set("Content-Length", fmt.Sprint(contentLen))
	headers.Set("Connection", "close")
	headers.Set("Content-Type", "text/plain")

	return *headers
}

func WriteHeaders(w io.Writer, headers headers.Headers) error  {
	var msg string
	for k, v := range headers {
		msg += fmt.Sprintf("%s: %s\r\n", k, v)
	}

	_, err := fmt.Fprintf(w, "%s\r\n", msg) // finish headers
	if err != nil {
		return fmt.Errorf("Failed to write headers: %w", err)
	}
	return nil
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

type HandlerError struct {
	StatusCode StatusCode
	Message    string
}

func NewHandlerErr(statusCode StatusCode) HandlerError {
	return HandlerError{
		StatusCode: statusCode,
		Message: statusCode.String(),
	}
}

func (e *HandlerError) Write(w io.Writer) {
	WriteStatusLine(w, e.StatusCode)
	WriteHeaders(w, GetDefaultHeaders(0))
}

func WriteError(w io.Writer, statusCode StatusCode) {
	WriteStatusLine(w, statusCode)
	WriteHeaders(w, GetDefaultHeaders(0))
}
