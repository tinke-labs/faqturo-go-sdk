package faqturo

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
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
	apiServer, err := normalizeServer(server)
	if err != nil {
		return nil, err
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
	return NewClientWithResponses(apiServer, append(defaults, options...)...)
}

// ValidateXML validates a received fiscal XML without issuing a document.
//
// This is the ergonomic SDK entry point for the generated ValidateXml
// operation. It returns the typed validation result instead of exposing the
// generated HTTP response wrapper.
func (c *ClientWithResponses) ValidateXML(ctx context.Context, xml []byte, reqEditors ...RequestEditorFn) (*XmlValidationResponse, error) {
	response, err := c.ValidateXml(ctx, nil, XmlValidationRequest{Xml: string(xml)}, reqEditors...)
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, errors.New("faqturo: validate XML returned no response")
	}
	if response.JSON200 == nil {
		if apiErr := ErrorFromResponse(response.HTTPResponse, response.Body); apiErr != nil {
			return nil, apiErr
		}
		return nil, fmt.Errorf("faqturo: validate XML returned HTTP %d", response.StatusCode())
	}
	return response.JSON200, nil
}

// normalizeServer accepts Faqturo's public host and resolves the API context
// path used by the SDK. Supplying an explicit path keeps that path unchanged,
// so reverse proxies and self-hosted deployments remain configurable.
func normalizeServer(server string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(server))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", errors.New("faqturo: API server URL is required")
	}
	if parsed.Path == "" || parsed.Path == "/" {
		parsed.Path = "/api"
	}
	return strings.TrimRight(parsed.String(), "/"), nil
}
