package faqturo

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAPIKeyClientAddsAuthenticationAndVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != "/api/catalogs/tax-codes" {
			t.Errorf("path = %q", got)
		}
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
	response, err := client.GetTaxCodes(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if response.JSON200 == nil {
		t.Fatal("expected typed JSON200 response")
	}
}

func TestAPIKeyClientExposesRawOperation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/catalogs/tax-codes" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client, err := NewAPIKeyClient(server.URL, "secret")
	if err != nil {
		t.Fatal(err)
	}
	response, err := client.Raw().GetTaxCodesRaw(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", response.StatusCode)
	}
}

func TestAPIKeyClientPreservesExplicitAPIPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != "/api/catalogs/tax-codes" {
			t.Errorf("path = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client, err := NewAPIKeyClient(server.URL+"/api", "secret")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetTaxCodes(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
}

func TestValidateXMLReturnsTypedResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/documents/validate-xml" {
			t.Errorf("request = %s %s", r.Method, r.URL.Path)
		}
		var request XMLValidationRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decoding request: %v", err)
		}
		if request.Xml != "<FacturaElectronica/>" {
			t.Errorf("xml = %q", request.Xml)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"valid":true,"issues":[]}`))
	}))
	defer server.Close()

	client, err := NewAPIKeyClient(server.URL, "secret")
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.ValidateXML(context.Background(), []byte("<FacturaElectronica/>"))
	if err != nil {
		t.Fatal(err)
	}
	if result == nil || !result.Valid {
		t.Fatalf("unexpected validation result: %#v", result)
	}
}

func TestStableOperationNamesAreGenerated(t *testing.T) {
	client := &ClientWithResponses{}
	_ = client.UpdateCashRegisterSequence
	_ = client.GetClientExoneration
	_ = client.GetTaxAuthorityExoneration
}

func TestAPIError(t *testing.T) {
	err := ErrorFromResponse(&http.Response{StatusCode: 422}, []byte(`{"message":"invalid","metadata":{"field":"amount"}}`))
	apiErr, ok := err.(*APIError)
	if !ok || apiErr.StatusCode != 422 || apiErr.Message != "invalid" {
		t.Fatalf("unexpected error: %#v", err)
	}
}

func TestAPIKeyClientRejectsInvalidServer(t *testing.T) {
	if _, err := NewAPIKeyClient("not a URL", "secret"); err == nil {
		t.Fatal("expected invalid server error")
	}
}
