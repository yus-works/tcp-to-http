package request

import (
	"fmt"
	"io"
)

type Request struct {
	RequestLine RequestLine
	state       parserState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type parserState string

const (
	StateInit parserState = "init"
	StateDone parserState = "done"
)

func (r *Request) parse(data []byte) (int, error) {
	for {
		switch r.state {
		case StateInit:
			rl, n, err := parseRequestLine(data)
			if err != nil {
				return 0, err
			}

			if n > 0 {
				r.RequestLine = *rl
				r.state = StateDone
			}
		
			// TODO: this will probably change when doing headers
			return n, nil

		case StateDone:
			return 0, nil
		}
	}
}

func (r *Request) done() bool {
	return r.state == StateDone
}

func newRequest() *Request {
	return &Request{
		state: StateInit,
	}
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()

	// TODO: add buffer resizing
	buf := make([]byte, 1024)

	// this indexes the last byte in the buf that stores data
	dataEnd := 0
	for !request.done() {

		// this just keeps reading into the buffer
		readN, readErr := reader.Read(buf[dataEnd:])

		dataEnd += readN

		parsedN, parseErr := request.parse(buf[:dataEnd])
		if parseErr != nil {
			return nil, parseErr
		}

		// when it returns non zero, it means it parsed a valid line
		// so we can just clear out the data that it parsed because we dont
		// need it anymore
		if parsedN > 0 {
			// since the parsed line might end before the end of
			// the latest read chunk, we copy anything that is left
			// after the length the parser says it consumed and copy it
			// to the start because that might be the start of another line

			copy(buf, buf[parsedN:dataEnd])
			dataEnd -= parsedN
		}

		if readErr != nil {
			if readErr == io.EOF {
				if parsedN == 0 && dataEnd > 0 {
					// have unparsed data at EOF
					return nil, fmt.Errorf("incomplete data at EOF")
				}
				if request.done() {
					break
				}

				// if not done but hit EOF
				return nil, io.ErrUnexpectedEOF
			}
			return nil, readErr
		}

		// keep reading and trying to parse until parse() returns non zero or
		// read errors
	}

	return request, nil
}
