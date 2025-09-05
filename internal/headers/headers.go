package headers

import (
	"bytes"
	"fmt"
)

type Headers map[string]string

var CRLF = []byte("\r\n")
var SP = byte(' ')

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, CRLF)
	if idx == -1 {
		return 0, false, nil
	}

	// if data is just crlf, all headers have been parsed
	if len(data) == len(CRLF) {
		return len(CRLF), true, nil
	}

	line := data[:idx]
	read := idx + len(CRLF)

	colonIdx := bytes.Index(line, []byte(":"))

	if colonIdx > 0 {
		if line[colonIdx - 1] == SP {
			return 0, false, fmt.Errorf(
				"Invalid format: key and ':' should not have a space between",
			)
		}
	} else {
		return 0, false, fmt.Errorf("Missing ':' in header line")
	}

	kv := bytes.TrimSpace(line)

	parts := bytes.SplitN(kv, []byte(":"), 2)
	key := string(parts[0])
	val := string(bytes.TrimSpace(parts[1]))

	h[key] = val

	return read, false, nil
}

func NewHeaders() Headers {
	return Headers{}
}
