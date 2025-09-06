package headers

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

type Headers map[string]string

var CRLF = []byte("\r\n")
var SP = byte(' ')
var isValidKey = regexp.MustCompile("^[a-zA-Z0-9!#$%&'*+-.^_`|~]+$")

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

	if !isValidKey.Match(parts[0]) {
		return "", "", fmt.Errorf("Invalid key format")
	}

	key := string(bytes.ToLower(parts[0]))
	val := string(bytes.TrimSpace(parts[1]))

	return key, val, nil
}

func (h Headers) Get(k string) string {
	return h[strings.ToLower(k)]
}

func (h Headers) Set(k, v string) {
	key := strings.ToLower(k)

	if oldVal, ok := h[key]; !ok { // not found
		h[key] = v

	} else { // found
		if v == oldVal {
			return

		} else {
			newVal := strings.Join([]string{oldVal, v}, ", ")
			h[key] = newVal
		}
	}
}

func (h Headers) Parse(data []byte) (int, bool, error) {
	read := 0
	done := false

	line := data

	// TODO: is this loop really necessary?
	for {
		line = data[read:]

		idx := bytes.Index(line, CRLF)
		if idx == -1 {
			break
		}

		// if data is just crlf, all headers have been parsed
		if len(line) == len(CRLF) {
			done = true
			break
		}

		line = line[:idx]

		k, v, err := parseHeader(line)
		if err != nil {
			return 0, done, err
		}

		read += idx + len(CRLF)

		h.Set(k, v)
	}

	return read, done, nil
}

func NewHeaders() *Headers {
	return &Headers{}
}
