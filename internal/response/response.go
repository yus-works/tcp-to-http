package response

import (
	"fmt"
	"io"
	"log"
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

func WriteStatusLine(w io.Writer, statusCode StatusCode) {
	_, err := fmt.Fprintf(w,
		"HTTP/1.1 %d %s\r\n",
		statusCode, statusCode.String(),
	)
	if err != nil {
		log.Println("Failed to write status line:", err)
	}
}
