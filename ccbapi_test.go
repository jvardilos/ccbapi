package ccbapi

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"
)

type mockClient struct {
	doFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockClient) Do(req *http.Request) (*http.Response, error) {
	return m.doFunc(req)
}

func TestCall_Success(t *testing.T) {
	client := &mockClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(`{"success":true}`)),
			}, nil
		},
	}

	token := &Token{AccessToken: "token", ExpiresIn: 9999999999}
	creds := &Credentials{}
	body, err := call("GET", "test", token, creds, client)
	if err != nil || string(body) != `{"success":true}` {
		t.Fatalf("expected success, got error: %v, body: %s", err, string(body))
	}
}

func TestCall_RequestError(t *testing.T) {
	client := &mockClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("request failed")
		},
	}
	token := &Token{AccessToken: "token", ExpiresIn: 9999999999}
	creds := &Credentials{}
	_, err := call("GET", "test", token, creds, client)
	if err == nil || err.Error() != "request failed" {
		t.Fatalf("expected request error, got: %v", err)
	}
}

func TestCall_BadStatus(t *testing.T) {
	client := &mockClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 401,
				Status:     "401 Unauthorized",
				Body:       io.NopCloser(bytes.NewBufferString("unauthorized")),
			}, nil
		},
	}
	token := &Token{AccessToken: "token", ExpiresIn: 9999999999}
	creds := &Credentials{}
	_, err := call("GET", "test", token, creds, client)
	if err == nil || err.Error() != "status: 401 Unauthorized, body unauthorized" {
		t.Fatalf("expected unauthorized error, got: %v", err)
	}
}

func TestCall_BodyReadError(t *testing.T) {
	client := &mockClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(badReader{}),
			}, nil
		},
	}
	token := &Token{AccessToken: "token", ExpiresIn: 9999999999}
	creds := &Credentials{}
	_, err := call("GET", "test", token, creds, client)
	if err == nil || err.Error() != "forced read error" {
		t.Fatalf("expected body read error, got: %v", err)
	}
}

type badReader struct{}

func (b badReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("forced read error")
}

func (b badReader) Close() error {
	return nil
}
