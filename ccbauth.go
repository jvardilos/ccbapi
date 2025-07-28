package ccbapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

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

/**
 * TODO: write tests and make it mockable
**/
func Auth(c *Credentials) (*Token, error) {
	codeChan := make(chan string, 1)
	c.RedirectURI = "http://localhost:8080/callback"

	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		return nil, fmt.Errorf("failed to bind listener: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "No code found", http.StatusBadRequest)
			return
		}
		fmt.Fprintln(w, "Authorization code received. You may now return to your terminal.")
		codeChan <- code
	})

	srv := &http.Server{Handler: mux}

	go func() {
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Printf("OAuth server error: %v", err)
		}
	}()

	browserAuth(c)

	select {
	case code := <-codeChan:
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
		c.Code = code
		return authorize(c)
	case <-time.After(120 * time.Second):
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
		return nil, fmt.Errorf("timed out waiting for OAuth authorization")
	}
}

/**
 * TODO: write tests and make it mockable
**/
func browserAuth(c *Credentials) {
	authURL := fmt.Sprintf("https://oauth.ccbchurch.com/oauth/authorize?response_type=code&client_id=%s&redirect_uri=%s",
		url.QueryEscape(c.Client),
		url.QueryEscape(c.RedirectURI),
	)

	if err := openBrowser(authURL); err != nil {
		fmt.Printf("Please open this URL manually: %s\n", authURL)
	} else {
		fmt.Println("Your browser has been opened to authorize the app.")
	}
}

func authorize(c *Credentials) (*Token, error) {

	tokenURL := "https://api.ccbchurch.com/oauth/token"

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", c.Code)
	form.Set("redirect_uri", c.RedirectURI)
	form.Set("subdomain", c.Subdomain)

	req, err := http.NewRequest("POST", tokenURL, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	req.SetBasicAuth(c.Client, c.Secret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var body bytes.Buffer
		body.ReadFrom(resp.Body)
		return nil, fmt.Errorf("token request failed: %s", body.String())
	}

	var token Token
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	// update token expiration time
	token.ExpiresIn = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second).Unix()

	return &token, nil
}

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	default:
		return exec.Command("xdg-open", url).Start()
	}
}

/**
 * Check if the token is still valid.
 * If the token is within 30 seconds of expiration, it is considered invalid.
 *
 * TODO: write tests and make it mockable
**/
func isAuthorized(t *Token) bool {
	now := time.Now().Unix()
	exp := t.ExpiresIn - 30
	return now < exp
}

/**
 * TODO: write tests and make it mockable
**/
func refresh(t *Token, c *Credentials) error {

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", t.RefreshToken)

	req, err := http.NewRequest("POST", "https://api.ccbchurch.com/oauth/token", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/vnd.ccbchurch.v2+json")
	req.SetBasicAuth(c.Client, c.Secret)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token refresh failed: %s %w", body, err)
	}

	var raw Token
	if err := json.Unmarshal(body, &raw); err != nil {
		return fmt.Errorf("decoding refresh response: %w", err)
	}

	t.AccessToken = raw.AccessToken
	t.RefreshToken = raw.RefreshToken
	t.ExpiresIn = time.Now().Add(time.Duration(raw.ExpiresIn) * time.Second).Unix()

	return nil
}
