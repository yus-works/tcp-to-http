package request

import (
	"fmt"
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
	consumed := 0
	// TODO: is this loop doing anything
	for {
		switch r.state {
		case StateInit:
			rl, n, err := parseRequestLine(data[consumed:])
			if err != nil {
				return 0, err
			}

			if n > 0 {
				r.RequestLine = *rl
				r.state = StateHeaders

				consumed += n
				continue // continue to process if theres more data
			}

			return consumed, nil // return if no parse

		case StateHeaders:
			n, done, err := r.Headers.Parse(data[consumed:])
			if err != nil {
				return consumed, err
			}

			if done {
				r.state = StateBody

				// done means CRLF at start of buf
				// so += 2 to skip those two bytes
				consumed += 2
				consumed += n

				continue
			}

			consumed += n

			if n == 0 {
				return consumed, nil
			}

		case StateBody:
			clen := r.Headers.Get("content-length")
			if clen == "" {
				r.state = StateDone
				return consumed, nil
			}

			ln, err := strconv.Atoi(clen)
			if err != nil {
				return consumed, err
			}

			remaining := ln - len(r.Body)
			available := len(data) - consumed
			toRead := min(remaining, available)

			r.Body = append(r.Body, data[consumed:consumed+toRead]...)
			consumed += toRead

			if len(r.Body) == ln {
				r.state = StateDone
			}
			return consumed, nil

		case StateDone:
			return consumed, nil

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

		if readN > 0 {
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
		}

		if readErr == io.EOF {
			// try to parse any remaining data
			if dataEnd > 0 {
				_, parseErr := request.parse(buf[:dataEnd])
				if parseErr != nil {
					return nil, parseErr
				}
			}

			// check if body is done
			if request.state == StateBody {
				if cl := request.Headers.Get("content-length"); cl != "" {
					ln, _ := strconv.Atoi(cl)
					if len(request.Body) < ln {
						return nil, fmt.Errorf("incomplete body: expected %d bytes, got %d", ln, len(request.Body))
					}
				}
				request.state = StateDone
			}

			if request.done() {
				break
			}

			return nil, fmt.Errorf("unexpected EOF in state %s", request.state)
		}

		if readErr != nil {
			return nil, readErr
		}

		// keep reading and trying to parse until parse() returns non zero or
		// read errors
	}

	return request, nil
}
