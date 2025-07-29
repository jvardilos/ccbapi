package ccbapi

import "net/http"

func Authorize(c *Credentials) (*Token, error) {
	return authorize(c, http.DefaultClient)
}

func Call(method, query string, t *Token, c *Credentials) ([]byte, error) {
	return call(method, query, t, c, http.DefaultClient)
}
