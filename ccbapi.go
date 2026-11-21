package ccbapi

import (
	"io"
	"net/http"
)

/**
 * TODO: write tests and make it mockable
**/
func call(method, query string, body io.Reader, t *Token, c *Credentials, client HTTPClient) ([]byte, error) {
	logger.Debug("calling ccb api", "method", method, "query", query)

	if !isAuthorized(t) {
		if err := refresh(t, c, client); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, "https://api.ccbchurch.com/"+query, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+t.AccessToken)
	req.Header.Set("Accept", "application/vnd.ccbchurch.v2+json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("ccb api request failed", "error", err)
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		logger.Error("ccb api request rejected", "status", resp.Status, "query", query)
		return nil, &APIError{StatusCode: resp.StatusCode, Status: resp.Status, Body: respBody}
	}

	return respBody, nil
}
