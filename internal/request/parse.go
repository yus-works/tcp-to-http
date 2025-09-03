package request

import (
	"bytes"
	"fmt"
)

var methods = map[string]struct{}{
	"GET":     {},
	"PUT":     {},
	"POST":    {},
	"DELETE":  {},
	"OPTIONS": {},
}

var versions = map[string]struct{}{
	"1.1": {},
}

var CRLF = []byte("\r\n")
var SP = []byte{' '}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	var reqLine RequestLine

	idx := bytes.Index(data, CRLF)
	if idx == -1 {
		return nil, 0, nil
	}

	line := data[:idx]
	read := idx + len(CRLF)

	parts := bytes.Split(line, []byte{' '})

	if len(parts) != 3 {
		return nil, 0, fmt.Errorf("Invalid number of request line parts")
	}

	method := string(parts[0])
	if _, ok := methods[method]; !ok {
		return nil, 0, fmt.Errorf("Request METHOD not found in allowed set")
	}

	reqLine.Method = method

	target := parts[1]
	if !isValidTarget.Match(target) {
		return nil, 0, fmt.Errorf("Request TARGET must follow [/].* (for now)")
	}

	reqLine.RequestTarget = string(target)

	versionToken := parts[2]
	versionParts := bytes.Split(versionToken, []byte{'/'})
	if string(versionParts[0]) != "HTTP" {
		return nil, 0, fmt.Errorf("Request type must be exactly 'HTTP'")
	}

	versionNum := string(versionParts[1])
	if _, ok := versions[versionNum]; !ok {
		return nil, 0, fmt.Errorf("Request version must be exactly '1.1' (for now)")
	}

	reqLine.HttpVersion = versionNum

	return &reqLine, read, nil
}
