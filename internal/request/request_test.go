package request

import (
	"io"
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
				"\r\n", numBytesPerRead: 3,
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
				"\r\n", numBytesPerRead: 3,
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
				"\r\n", numBytesPerRead: 3,
		},
	)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "POST", r.RequestLine.Method)
	assert.Equal(t, "/api/users", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// TODO: disabled test: PUT method with query parameters
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

	// Test: DELETE method
	r, err = RequestFromReader(
		&chunkReader{
			data: "DELETE /users/123 HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"Authorization: Bearer token123\r\n" +
				"\r\n", numBytesPerRead: 3,
		},
	)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "DELETE", r.RequestLine.Method)
	assert.Equal(t, "/users/123", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: HTTP/1.0 version // TODO: only allowing 1.1 for now
	_, err = RequestFromReader(
		&chunkReader{
			data: "\r\n" +
				"Host: localhost:42069\r\n" +
				"\r\n", numBytesPerRead: 3,
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
				"\r\n", numBytesPerRead: 3,
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
			numBytesPerRead: 3,
		},
	)
	require.Error(t, err)

	// Test: Too many parts in request line
	_, err = RequestFromReader(
		&chunkReader{
			data: "GET /path HTTP/1.1 extra\r\n" +
				"Host: localhost:42069\r\n" + "\r\n",
			numBytesPerRead: 3,
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
			numBytesPerRead: 3,
		},
	)
	require.Error(t, err)

	// Test: Invalid HTTP version prefix
	_, err = RequestFromReader(
		&chunkReader{
			data: "GET /path HTTPS/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"\r\n",
			numBytesPerRead: 3,
		},
	)
	require.Error(t, err)

	// Test: Method with lowercase characters
	_, err = RequestFromReader(
		&chunkReader{
			data: "get /path HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"\r\n",
			numBytesPerRead: 3,
		},
	)
	require.Error(t, err)

	// TODO: disabled test: only allowing 1.1 for now
	// r, err = RequestFromReader(
	// 	&chunkReader{
	// 		data:"GET /legacy HTTP/1.0\r\n" +
	// 		"Host: localhost:42069\r\n" +
	// 		"\r\n",	numBytesPerRead: 3,
	// 	},
	// )
	// require.NoError(t, err)
	// require.NotNil(t, r)
	// assert.Equal(t, "GET", r.RequestLine.Method)
	// assert.Equal(t, "/legacy", r.RequestLine.RequestTarget)
	// assert.Equal(t, "1.0", r.RequestLine.HttpVersion)

	// TODO: disabled test: not allowing method versioning yet
	// r, err = RequestFromReader(
	// 	&chunkReader{
	// 		data:"PATCH-V1.2 /path HTTP/1.1\r\n" +
	// 		"Host: localhost:42069\r\n" +
	// 		"\r\n",	numBytesPerRead: 3,
	// 	}
	// )
	// require.NoError(t, err)
	// require.NotNil(t, r)
	// assert.Equal(t, "PATCH-V1.2", r.RequestLine.Method)

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
				"\r\n", numBytesPerRead: 3,
		},
	)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/path%20with%20spaces", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// TODO: Disabled test: Long request target
	// r, err = RequestFromReader(
	// 	&chunkReader{
	// 		data:"GET /very/long/path/with/many/segments/and/query?param1=value1&param2=value2&param3=value3 HTTP/1.1\r\n" +
	// 		"Host: localhost:42069\r\n" +
	// 		"\r\n",	numBytesPerRead: 3,
	// 	},
	// )
	// require.NoError(t, err)
	// require.NotNil(t, r)
	// assert.Equal(t, "GET", r.RequestLine.Method)
	// assert.Equal(t, "/very/long/path/with/many/segments/and/query?param1=value1&param2=value2&param3=value3", r.RequestLine.RequestTarget)
	// assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: Empty method
	_, err = RequestFromReader(
		&chunkReader{
			data: " /path HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"\r\n", numBytesPerRead: 3,
		},
	)
	require.Error(t, err)

	// Test: Empty request target
	_, err = RequestFromReader(
		&chunkReader{
			data: "GET  HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"\r\n", numBytesPerRead: 3,
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
}
