package ccbapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"time"
)

/**
 * TODO: write tests and make it mockable
**/
func authorize(c *Credentials, client HTTPClient) (*Token, error) {
	codeChan := make(chan string, 1)
	c.RedirectURI = "http://localhost:8080/callback"

	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		return nil, err
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

	go func() error {
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			return err
		}

		return nil
	}()

	browserAuth(c)

	select {
	case code := <-codeChan:
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
		c.Code = code
		return step2(c, client)
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

func step2(c *Credentials, client HTTPClient) (*Token, error) {

	tokenURL := "https://api.ccbchurch.com/oauth/token"

	body, err := json.Marshal(map[string]string{
		"grant_type":    "authorization_code",
		"subdomain":     c.Subdomain,
		"client_id":     c.Client,
		"client_secret": c.Secret,
		"code":          c.Code,
		"redirect_uri":  c.RedirectURI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to encode token request: %w", err)
	}

	req, err := http.NewRequest("POST", tokenURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

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
func refresh(t *Token, c *Credentials, client HTTPClient) error {

	body, err := json.Marshal(map[string]string{
		"grant_type":    "refresh_token",
		"subdomain":     c.Subdomain,
		"client_id":     c.Client,
		"client_secret": c.Secret,
		"refresh_token": t.RefreshToken,
	})
	if err != nil {
		return fmt.Errorf("failed to encode refresh request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.ccbchurch.com/oauth/token", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token refresh failed: %s", respBody)
	}

	var raw Token
	if err := json.Unmarshal(respBody, &raw); err != nil {
		return fmt.Errorf("decoding refresh response: %w", err)
	}

	t.AccessToken = raw.AccessToken
	t.RefreshToken = raw.RefreshToken
	t.ExpiresIn = time.Now().Add(time.Duration(raw.ExpiresIn) * time.Second).Unix()

	return nil
}
