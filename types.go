package ccbapi

import "net/http"

type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

type Credentials struct {
	Code        string
	RedirectURI string
	Subdomain   string
	Client      string
	Secret      string
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}
