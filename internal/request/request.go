package request

import (
	"io"
	"strconv"

	"github.com/yus-works/tcp-to-http/internal/headers"
)

type Request struct {
	RequestLine RequestLine
	state       parserState
	Headers     *headers.Headers
	Body        []byte
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type parserState string

const (
	StateInit    parserState = "init"
	StateDone    parserState = "done"
	StateHeaders parserState = "headers"
	StateBody    parserState = "body"
)

func (r *Request) parse(data []byte) (int, error) {
	// TODO: this should be called consumedN because it only says
	// how many bytes were parsed when a successful flush happened
	read := 0
	// TODO: is this loop doing anything
	for {
		switch r.state {
		case StateInit:
			rl, n, err := parseRequestLine(data[read:])
			if err != nil {
				return 0, err
			}

			if n > 0 {
				r.RequestLine = *rl
				r.state = StateHeaders
			}

			read += n

			return read, nil

		case StateHeaders:
			n, done, err := r.Headers.Parse(data[read:])
			if err != nil {
				return 0, err
			}

			if done {
				r.state = StateBody

				// done means CRLF at start of buf
				// so += 2 to skip those two bytes
				read += 2
			}

			read += n

			if n == 0 {
				return read, nil
			}

			// NOTE: important
			return read, nil

		case StateBody:
			clen := r.Headers.Get("content-length")
			if clen == "" {
				r.state = StateDone
				return read, nil
			}

			r.Body = append(r.Body, data[read:]...)
			read += len(data) - read

			ln, err := strconv.Atoi(clen)
			if err != nil {
				return 0, nil
			}

			// TODO: probably terrible for perf
			if len(r.Body) == ln {
				r.state = StateDone
			}

			return read, nil

		case StateDone:
			return 0, nil

		default:
			panic("ayo what")
		}
	}
}

func (r *Request) done() bool {
	return r.state == StateDone
}

func newRequest() *Request {
	return &Request{
		state:   StateInit,
		Headers: headers.NewHeaders(),
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

		// TODO: horrible
		if readErr == io.EOF {
			// if EOF but in body parsing stage, try parsing body
			if request.state == StateBody {
				// if EOF and nothing left to read, error
				if readN == 0 {
					// this has to be here I think
					// (so maybe take it out of .parse?)
					if cl := request.Headers.Get("content-length"); cl == "" {
						break
					}
					return nil, readErr
				}

				continue
			}

			if request.done() {
				break
			}
		}

		if readErr != nil {
			return nil, readErr
		}

		// keep reading and trying to parse until parse() returns non zero or
		// read errors
	}

	return request, nil
}
