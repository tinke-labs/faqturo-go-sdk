package faqturo

import (
	"context"
	"errors"
	"net/http"
	"time"
)

// WithTimeout configures the standard HTTP client timeout. Apply it after
// WithHTTPClient when both options are used.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(client *Client) error {
		if timeout <= 0 {
			return errors.New("faqturo: timeout must be positive")
		}
		if client.Client == nil {
			client.Client = &http.Client{Timeout: timeout}
			return nil
		}
		httpClient, ok := client.Client.(*http.Client)
		if !ok {
			return errors.New("faqturo: timeout requires *http.Client")
		}
		clone := *httpClient
		clone.Timeout = timeout
		client.Client = &clone
		return nil
	}
}

// WithAPIVersion overrides the default v1 API-Version header.
func WithAPIVersion(version string) ClientOption {
	return func(client *Client) error {
		if version == "" {
			return errors.New("faqturo: API version is required")
		}
		client.RequestEditors = append(client.RequestEditors, func(_ context.Context, req *http.Request) error {
			req.Header.Set("API-Version", version)
			return nil
		})
		return nil
	}
}

// WithRequestEditor is the idiomatic alias for the generated editor option.
func WithRequestEditor(editor RequestEditorFn) ClientOption {
	if editor == nil {
		return func(*Client) error { return errors.New("faqturo: nil request editor") }
	}
	return WithRequestEditorFn(editor)
}

func NewAPIKeyClient(server, apiKey string, options ...ClientOption) (*ClientWithResponses, error) {
	if apiKey == "" {
		return nil, errors.New("faqturo: API key is required")
	}
	auth := func(_ context.Context, req *http.Request) error {
		req.Header.Set("X-API-KEY", apiKey)
		req.Header.Set("API-Version", "v1")
		return nil
	}
	defaults := []ClientOption{
		WithHTTPClient(&http.Client{Timeout: 30 * time.Second}),
		WithRequestEditorFn(auth),
	}
	return NewClientWithResponses(server, append(defaults, options...)...)
}
