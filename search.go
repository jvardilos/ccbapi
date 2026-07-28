package ccbapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// IndividualActivity is a subset of CCB's individual search result fields
// relevant to engagement tracking: identity plus last-known interaction
// dates, obtained without a separate per-individual metrics lookup.
type IndividualActivity struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	Email            string `json:"email"`
	LastAttendedDate string `json:"last_attended_date"`
	LastGivingDate   string `json:"last_giving_date"`
}

type individualSearchRequest struct {
	Filters map[string]any `json:"filters"`
}

// SearchIndividuals runs CCB's advanced individual search and returns one
// page of individuals along with their last-attended and last-giving dates
// in a single request, in place of one request per individual per metric.
func SearchIndividuals(t *Token, c *Credentials, page, perPage int) ([]IndividualActivity, error) {
	return searchIndividuals(t, c, http.DefaultClient, page, perPage)
}

func searchIndividuals(t *Token, c *Credentials, client HTTPClient, page, perPage int) ([]IndividualActivity, error) {
	payload, err := json.Marshal(individualSearchRequest{Filters: map[string]any{}})
	if err != nil {
		return nil, fmt.Errorf("encoding search request: %w", err)
	}

	query := fmt.Sprintf("search/individuals/results?per_page=%d&page=%d", perPage, page)
	body, err := call("POST", query, bytes.NewReader(payload), t, c, client)
	if err != nil {
		return nil, err
	}

	var results []IndividualActivity
	if err := json.Unmarshal(body, &results); err != nil {
		return nil, fmt.Errorf("decoding search results: %w", err)
	}

	return results, nil
}
