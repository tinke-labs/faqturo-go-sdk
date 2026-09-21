package faqturo

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// APIError is the normalized ErrorResponse returned by Faqturo.
type APIError struct {
	StatusCode       int               `json:"status"`
	ErrorCode        string            `json:"error"`
	Message          string            `json:"message"`
	Path             string            `json:"path"`
	Details          string            `json:"details,omitempty"`
	ValidationErrors map[string]string `json:"validationErrors,omitempty"`
	Metadata         map[string]any    `json:"metadata,omitempty"`
	Body             []byte            `json:"-"`
	Provider         string            `json:"-"`
	Operation        string            `json:"-"`
	UpstreamCode     string            `json:"-"`
	RequestID        string            `json:"-"`
	Retryable        bool              `json:"-"`
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("faqturo: HTTP %d: %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("faqturo: HTTP %d", e.StatusCode)
}

func ErrorFromResponse(response *http.Response, body []byte) error {
	if response == nil || response.StatusCode < 400 {
		return nil
	}
	err := &APIError{StatusCode: response.StatusCode, Body: append([]byte(nil), body...)}
	_ = json.Unmarshal(body, err)
	err.StatusCode = response.StatusCode
	err.Provider = "faqturo"
	err.UpstreamCode = err.ErrorCode
	err.RequestID = firstNonEmpty(
		response.Header.Get("X-Request-ID"),
		response.Header.Get("X-Correlation-ID"),
		metadataString(err.Metadata, "requestId"),
		metadataString(err.Metadata, "requestID"),
	)
	err.Retryable = response.StatusCode == http.StatusRequestTimeout || response.StatusCode == http.StatusTooManyRequests || response.StatusCode >= 500
	return err
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func metadataString(metadata map[string]any, key string) string {
	if metadata == nil {
		return ""
	}
	value, ok := metadata[key]
	if !ok {
		return ""
	}
	text, _ := value.(string)
	return text
}
