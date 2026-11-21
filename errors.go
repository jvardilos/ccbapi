package ccbapi

import "fmt"

// APIError is returned by Call (and the helpers built on it) when CCB answers
// with a status other than 200. Callers can inspect it with errors.As to
// branch on StatusCode or Body instead of parsing the error message.
type APIError struct {
	StatusCode int    // numeric HTTP status, e.g. 400
	Status     string // full status line, e.g. "400 Bad Request"
	Body       []byte // raw response body
}

func (e *APIError) Error() string {
	return fmt.Sprintf("status: %s, body %s", e.Status, e.Body)
}
