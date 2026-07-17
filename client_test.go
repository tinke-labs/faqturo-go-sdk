package faqturo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAPIKeyClientAddsAuthenticationAndVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-API-KEY"); got != "secret" {
			t.Errorf("API key = %q", got)
		}
		if got := r.Header.Get("API-Version"); got != "v1" {
			t.Errorf("version = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()
	client, err := NewAPIKeyClient(server.URL, "secret", WithTimeout(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	response, err := client.GetTaxCodesWithResponse(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if response.JSON200 == nil {
		t.Fatal("expected typed JSON200 response")
	}
}

func TestStableOperationNamesAreGenerated(t *testing.T) {
	client := &ClientWithResponses{}
	_ = client.UpdateCashRegisterSequenceWithResponse
	_ = client.GetClientExonerationWithResponse
	_ = client.GetTaxAuthorityExonerationWithResponse
}

func TestAPIError(t *testing.T) {
	err := ErrorFromResponse(&http.Response{StatusCode: 422}, []byte(`{"message":"invalid","metadata":{"field":"amount"}}`))
	apiErr, ok := err.(*APIError)
	if !ok || apiErr.StatusCode != 422 || apiErr.Message != "invalid" {
		t.Fatalf("unexpected error: %#v", err)
	}
}
