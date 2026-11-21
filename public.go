package ccbapi

import (
	"io"
	"net/http"
)

func Authorize(c *Credentials) (*Token, error) {
	return authorize(c, http.DefaultClient)
}

func Call(method, query string, t *Token, c *Credentials) ([]byte, error) {
	return call(method, query, nil, t, c, http.DefaultClient)
}

// CallWithBody is like Call but sends body (sent as application/json) with
// the request, for POST and PUT. Any 2xx response is treated as success.
func CallWithBody(method, query string, body io.Reader, t *Token, c *Credentials) ([]byte, error) {
	return call(method, query, body, t, c, http.DefaultClient)
}
