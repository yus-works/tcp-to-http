package request

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network connection
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := min(cr.pos+cr.numBytesPerRead, len(cr.data))
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n

	return n, nil
}

func TestRequestLineParse(t *testing.T) {
	// Test: Good GET Request line
	r, err := RequestFromReader(
		&chunkReader{
			data: "GET / HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"User-Agent: curl/7.81.0\r\n" +
				"Accept: */*\r\n" +
				"\r\n", numBytesPerRead: 1,
		},
	)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: Good GET Request line with path
	r, err = RequestFromReader(
		&chunkReader{
			data: "GET /coffee HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"User-Agent: curl/7.81.0\r\n" +
				"Accept: */*\r\n" +
				"\r\n", numBytesPerRead: 2,
		},
	)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: Invalid number of parts in request line
	_, err = RequestFromReader(
		&chunkReader{
			data: "/coffee HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"User-Agent: curl/7.81.0\r\n" +
				"Accept: */*\r\n" +
				"\r\n", numBytesPerRead: 3,
		},
	)
	require.Error(t, err)

	// Test: POST method
	r, err = RequestFromReader(
		&chunkReader{
			data: "POST /api/users HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"Content-Type: application/json\r\n" +
				"\r\n", numBytesPerRead: 4,
		},
	)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "POST", r.RequestLine.Method)
	assert.Equal(t, "/api/users", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: DELETE method
	r, err = RequestFromReader(
		&chunkReader{
			data: "DELETE /users/123 HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"Authorization: Bearer token123\r\n" +
				"\r\n", numBytesPerRead: 1,
		},
	)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "DELETE", r.RequestLine.Method)
	assert.Equal(t, "/users/123", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: Missing method
	_, err = RequestFromReader(
		&chunkReader{
			data: "\r\n" +
				"Host: localhost:42069\r\n" +
				"\r\n", numBytesPerRead: 2,
		},
	)
	require.Error(t, err)

	// Test: Absolute URI in request target
	r, err = RequestFromReader(
		&chunkReader{
			data: "GET http://example.com/path HTTP/1.1\r\n" +
				"Host: example.com\r\n" +
				"\r\n", numBytesPerRead: 3,
		},
	)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "http://example.com/path", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: Asterisk-form request target (for OPTIONS)
	r, err = RequestFromReader(
		&chunkReader{
			data: "OPTIONS * HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"\r\n", numBytesPerRead: 4,
		},
	)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "OPTIONS", r.RequestLine.Method)
	assert.Equal(t, "*", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: Empty request line
	_, err = RequestFromReader(
		&chunkReader{
			data: "\r\n" +
				"Host: localhost:42069\r\n" +
				"\r\n",
			numBytesPerRead: 1,
		},
	)
	require.Error(t, err)

	// Test: Too many parts in request line
	_, err = RequestFromReader(
		&chunkReader{
			data: "GET /path HTTP/1.1 extra\r\n" +
				"Host: localhost:42069\r\n" + "\r\n",
			numBytesPerRead: 2,
		},
	)
	require.Error(t, err)

	// Test: Request line with only two parts
	_, err = RequestFromReader(
		&chunkReader{
			data: "GET /path\r\n" +
				"Host: localhost:42069\r\n" +
				"\r\n",
			numBytesPerRead: 3,
		},
	)
	require.Error(t, err)

	// Test: Invalid HTTP version format
	_, err = RequestFromReader(
		&chunkReader{
			data: "GET /path HTTP/2.0\r\n" +
				"Host: localhost:42069\r\n" +
				"\r\n",
			numBytesPerRead: 4,
		},
	)
	require.Error(t, err)

	// Test: Invalid HTTP version prefix
	_, err = RequestFromReader(
		&chunkReader{
			data: "GET /path HTTPS/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"\r\n",
			numBytesPerRead: 1,
		},
	)
	require.Error(t, err)

	// Test: Method with lowercase characters
	_, err = RequestFromReader(
		&chunkReader{
			data: "get /path HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"\r\n",
			numBytesPerRead: 2,
		},
	)
	require.Error(t, err)

	// Test: Request line without CRLF (only LF)
	_, err = RequestFromReader(
		&chunkReader{
			data: "GET /path HTTP/1.1\n" +
				"Host: localhost:42069\r\n" +
				"\r\n", numBytesPerRead: 3,
		},
	)
	require.Error(t, err)

	// Test: Request target with encoded characters
	r, err = RequestFromReader(
		&chunkReader{
			data: "GET /path%20with%20spaces HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"\r\n", numBytesPerRead: 4,
		},
	)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/path%20with%20spaces", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: Empty method
	_, err = RequestFromReader(
		&chunkReader{
			data: " /path HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"\r\n", numBytesPerRead: 1,
		},
	)
	require.Error(t, err)

	// Test: Empty request target
	_, err = RequestFromReader(
		&chunkReader{
			data: "GET  HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"\r\n", numBytesPerRead: 2,
		},
	)
	require.Error(t, err)

	// Test: Missing HTTP version
	_, err = RequestFromReader(
		&chunkReader{
			data: "GET /path \r\n" +
				"Host: localhost:42069\r\n" +
				"\r\n", numBytesPerRead: 3,
		},
	)
	require.Error(t, err)

	// TODO: enable these tests

	// Test: Long request target
	// r, err = RequestFromReader(
	// 	&chunkReader{
	// 		data:"GET /very/long/path/with/many/segments/and/query?param1=value1&param2=value2&param3=value3 HTTP/1.1\r\n" +
	// 		"Host: localhost:42069\r\n" +
	// 		"\r\n",	numBytesPerRead: 4,
	// 	},
	// )
	// require.NoError(t, err)
	// require.NotNil(t, r)
	// assert.Equal(t, "GET", r.RequestLine.Method)
	// assert.Equal(t, "/very/long/path/with/many/segments/and/query?param1=value1&param2=value2&param3=value3", r.RequestLine.RequestTarget)
	// assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: legacy version
	// r, err = RequestFromReader(
	// 	&chunkReader{
	// 		data:"GET /legacy HTTP/1.0\r\n" +
	// 		"Host: localhost:42069\r\n" +
	// 		"\r\n",	numBytesPerRead: 1,
	// 	},
	// )
	// require.NoError(t, err)
	// require.NotNil(t, r)
	// assert.Equal(t, "GET", r.RequestLine.Method)
	// assert.Equal(t, "/legacy", r.RequestLine.RequestTarget)
	// assert.Equal(t, "1.0", r.RequestLine.HttpVersion)

	// Test: versioned method
	// r, err = RequestFromReader(
	// 	&chunkReader{
	// 		data:"PATCH-V1.2 /path HTTP/1.1\r\n" +
	// 		"Host: localhost:42069\r\n" +
	// 		"\r\n",	numBytesPerRead: 2,
	// 	}
	// )
	// require.NoError(t, err)
	// require.NotNil(t, r)
	// assert.Equal(t, "PATCH-V1.2", r.RequestLine.Method)

	// Test: PUT method with query parameters
	// _, err = RequestFromReader(
	// 	&chunkReader{
	// 		data:"PUT /users/123?active=true HTTP/1.1\r\n" +
	// 		"Host: localhost:42069\r\n" +
	// 		"Content-Type: application/json\r\n" +
	// 		"\r\n",	numBytesPerRead: 3,
	// 	},
	// )
	// require.NoError(t, err)
	// require.NotNil(t, r)
	// assert.Equal(t, "PUT", r.RequestLine.Method)
	// assert.Equal(t, "/users/123?active=true", r.RequestLine.RequestTarget)
	// assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
}

func TestRequestParseHeaders(t *testing.T) {
	// Test: Standard Headers
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "localhost:42069", r.Headers.Get("host"))
	assert.Equal(t, "curl/7.81.0", r.Headers.Get("user-agent"))
	assert.Equal(t, "*/*", r.Headers.Get("accept"))

	// Test: Empty Headers
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, 0, len(*r.Headers))

	// Test: Malformed Header
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost localhost:42069\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.Error(t, err)

	// Test: Duplicate Headers
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nSet-Cookie: session=abc\r\nSet-Cookie: user=123\r\nSet-Cookie: theme=dark\r\n\r\n",
		numBytesPerRead: 5,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "session=abc, user=123, theme=dark", r.Headers.Get("set-cookie"))

	// Test: Case Insensitive Headers
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHOST: example.com\r\nContent-Type: text/html\r\n\r\n",
		numBytesPerRead: 4,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "example.com", r.Headers.Get("host"))
	assert.Equal(t, "text/html", r.Headers.Get("content-type"))

	// Test: Missing End of Headers
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost\r\nUser-Agent: test",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.Error(t, err)

	// Test: Headers with whitespace
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost:    spaced.com    \r\nAccept:  text/html  \r\n\r\n",
		numBytesPerRead: 4,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "spaced.com", r.Headers.Get("host"))
	assert.Equal(t, "text/html", r.Headers.Get("accept"))

	// Test: Single character chunks with headers
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nX-Custom: value\r\n\r\n",
		numBytesPerRead: 1,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "value", r.Headers.Get("x-custom"))

	// Test: Large header value
	largeValue := strings.Repeat("a", 1000)
	reader = &chunkReader{
		data:            fmt.Sprintf("GET / HTTP/1.1\r\nX-Large: %s\r\n\r\n", largeValue),
		numBytesPerRead: 10,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, largeValue, r.Headers.Get("x-large"))

	// Test: Invalid header characters
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nInvalid-©: value\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.Error(t, err)
}

func TestBodyParse(t *testing.T) {
	// Test: Standard Body
	reader := &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 13\r\n" +
			"\r\n" +
			"hello world!\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "hello world!\n", string(r.Body))

	// Test: Empty Body, 0 reported content length
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 0\r\n" +
			"\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "", string(r.Body))

	// Test: Empty Body, no reported content length
	reader = &chunkReader{
		data: "GET / HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "", string(r.Body))

	// Test: Body shorter than reported content length
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 20\r\n" +
			"\r\n" +
			"partial content",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.Error(t, err)

	// Test: No Content-Length but Body Exists
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"\r\n" +
			"body without length",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "", string(r.Body)) // Body should be empty since no Content-Length

	// Test: Large body with exact content length
	largeBody := "This is a larger body content that spans multiple chunks!"
	reader = &chunkReader{
		data: "POST /data HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			fmt.Sprintf("Content-Length: %d\r\n", len(largeBody)) +
			"\r\n" +
			largeBody,
		numBytesPerRead: 5,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, largeBody, string(r.Body))

	// Test: Body with single byte chunks
	reader = &chunkReader{
		data: "POST /api HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 5\r\n" +
			"\r\n" +
			"12345",
		numBytesPerRead: 1,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "12345", string(r.Body))

	// Test: Binary body content
	reader = &chunkReader{
		data: "POST /upload HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 4\r\n" +
			"\r\n" +
			"\x00\x01\x02\x03",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, []byte{0x00, 0x01, 0x02, 0x03}, r.Body)
}
