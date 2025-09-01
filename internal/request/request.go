package request

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"regexp"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

var methods = map[string]struct{}{
	"GET": {},
	"PUT": {},
	"POST": {},
	"DELETE": {},
	"OPTIONS": {},
}

var versions = map[string]struct{}{
	"1.1": {},
}

// TODO: use better validation for this
var isValidTarget = regexp.MustCompile("[*/][-_a-zA-Z0-9]*")

func parseRequestLine(line []byte) (*RequestLine, error) {
	var reqLine RequestLine
	var err error

	parts := bytes.Split(line, []byte{' '})

	if len(parts) != 3 {
		err = fmt.Errorf("Request line must contain exactly 3 parts")
		log.Println("Invalid number of request line parts: ", err)
		return nil, err
	}

	method := string(parts[0])
	if _, ok := methods[method]; !ok {
		err = fmt.Errorf("Request METHOD not found in allowed set")
		log.Println("Invalid METHOD: ", err)
		return nil, err
	}

	reqLine.Method = method

	target := parts[1]
	if !isValidTarget.Match(target) {
		err = fmt.Errorf("Request TARGET must follow [/].* (for now)")
		log.Println("Invalid TARGET: ", err)
		return nil, err
	}

	reqLine.RequestTarget = string(target)

	versionToken := parts[2]
	versionParts := bytes.Split(versionToken, []byte{'/'})
	if string(versionParts[0]) != "HTTP" {
		err = fmt.Errorf("Request type must be exactly 'HTTP'")
		log.Println("Invalid TYPE: ", err)
		return nil, err
	}

	versionNum := string(versionParts[1])
	if _, ok := versions[versionNum]; !ok {
		err = fmt.Errorf("Request version must be exactly '1.1' (for now)")
		log.Println("Invalid VERSION: ", err)
		return nil, err
	}

	reqLine.HttpVersion = versionNum

	return &reqLine, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	msg, err := io.ReadAll(reader)
	if err != nil {
		log.Println("Failed to read http message: ", err)
		return nil, err
	}

	line := bytes.Split(msg, []byte{'\r', '\n'})[0]

	reqLine, err := parseRequestLine(line)
	if err != nil {
		log.Println("Failed to parse request line: ", err)
		return nil, err
	}

	return &Request{
		RequestLine: *reqLine,
	}, nil
}
