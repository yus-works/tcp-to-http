package headers

import (
	"bytes"
	"fmt"
)

type Headers map[string]string

var CRLF = []byte("\r\n")
var SP = byte(' ')

func parseHeader(fieldLine []byte) (string, string, error) {
	colonIdx := bytes.Index(fieldLine, []byte(":"))

	if colonIdx > 0 {
		if fieldLine[colonIdx-1] == SP {
			return "", "", fmt.Errorf(
				"Invalid format: key and ':' should not have a space between",
			)
		}
	} else {
		return "", "", fmt.Errorf("Missing ':' in header line")
	}

	kv := bytes.TrimSpace(fieldLine)

	parts := bytes.SplitN(kv, []byte(":"), 2)
	key := string(parts[0])
	val := string(bytes.TrimSpace(parts[1]))

	return key, val, nil
}

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

	k, v, err := parseHeader(line)
	if err != nil {
		return 0, false, err
	}

	h[k] = v

	return read, false, nil
}

func NewHeaders() Headers {
	return Headers{}
}
