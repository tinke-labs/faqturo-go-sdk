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
	return err
}
