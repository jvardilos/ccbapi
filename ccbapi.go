package ccbapi

import (
	"fmt"
	"io"
	"net/http"
)

/**
 * TODO: write tests and make it mockable
**/
func Call(method, query string, t *Token, c *Credentials) ([]byte, error) {

	if !isAuthorized(t) {
		if err := refresh(t, c); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, "https://api.ccbchurch.com/"+query, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+t.AccessToken)
	req.Header.Set("Accept", "application/vnd.ccbchurch.v2+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status: %s, body %s", resp.Status, body)
	}

	return body, nil
}
