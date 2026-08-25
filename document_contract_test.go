package faqturo

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCreateInvoiceContract(t *testing.T) {
	const issuedAt = "2026-08-11T10:16:58.882370375-06:00"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/documents/invoice" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("X-API-KEY"); got != "secret" {
			t.Fatalf("X-API-KEY = %q", got)
		}
		if got := r.Header.Get("API-Version"); got != "v2" {
			t.Fatalf("API-Version = %q", got)
		}
		if got := r.Header.Get("Idempotency-Key"); got != "invoice-123" {
			t.Fatalf("Idempotency-Key = %q", got)
		}
		if got := r.Header.Get("Content-Type"); !strings.HasPrefix(got, "application/json") {
			t.Fatalf("Content-Type = %q", got)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if got := body["currency"]; got != "CRC" {
			t.Fatalf("currency = %#v", got)
		}
		items, ok := body["items"].([]any)
		if !ok || len(items) != 1 {
			t.Fatalf("items = %#v", body["items"])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":123,"issueDate":"`+issuedAt+`","createdAt":"`+issuedAt+`","updatedAt":"`+issuedAt+`","documentTotal":199.95}`)
	}))
	defer server.Close()

	client, err := NewAPIKeyClient(server.URL, "secret", WithAPIVersion("v2"))
	if err != nil {
		t.Fatal(err)
	}
	key := "invoice-123"
	currency := DocumentRequestCurrency("CRC")
	response, err := client.CreateInvoice(context.Background(), &CreateInvoiceParams{IdempotencyKey: &key}, DocumentRequest{
		Currency: &currency,
		Items: []DocumentItemRequest{{
			CabysCode:     "1234567890123",
			Description:   "Producto de prueba",
			LineNumber:    1,
			Quantity:      MustDecimal("1"),
			UnitOfMeasure: "Unid",
			UnitPrice:     MustDecimal("199.95"),
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if response.JSON200 == nil || response.JSON200.Id == nil || *response.JSON200.Id != 123 {
		t.Fatalf("unexpected response: %#v", response.JSON200)
	}
	if response.JSON200.IssueDate == nil || response.JSON200.IssueDate.Format(time.RFC3339Nano) != issuedAt {
		t.Fatalf("issueDate = %v", response.JSON200.IssueDate)
	}
}

func TestDocumentResponseRejectsTimestampWithoutOffset(t *testing.T) {
	timestampFields := []string{
		"issueDate", "paidAt", "paymentDueDate", "sentToHaciendaAt", "acceptedByHaciendaAt",
		"rejectedByHaciendaAt", "pdfGeneratedAt", "createdAt", "updatedAt",
	}
	for _, field := range timestampFields {
		t.Run(field, func(t *testing.T) {
			response := &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"` + field + `":"2026-08-11T10:16:58.882370375"}`)),
			}
			if _, err := ParseCreateInvoiceResponse(response); err == nil {
				t.Fatal("expected RFC3339 parsing error")
			}
		})
	}
}

func TestDocumentResponseParsesAllRFC3339TimestampFields(t *testing.T) {
	const timestamp = "2026-08-11T10:16:58.882370375-06:00"
	response := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body: io.NopCloser(strings.NewReader(`{
			"issueDate":"` + timestamp + `", "paidAt":"` + timestamp + `",
			"paymentDueDate":"` + timestamp + `", "sentToHaciendaAt":"` + timestamp + `",
			"acceptedByHaciendaAt":"` + timestamp + `", "rejectedByHaciendaAt":"` + timestamp + `",
			"pdfGeneratedAt":"` + timestamp + `", "createdAt":"` + timestamp + `",
			"updatedAt":"` + timestamp + `"
		}`)),
	}
	parsed, err := ParseCreateInvoiceResponse(response)
	if err != nil {
		t.Fatal(err)
	}
	for name, value := range map[string]*time.Time{
		"issueDate": parsed.JSON200.IssueDate, "paidAt": parsed.JSON200.PaidAt,
		"paymentDueDate": parsed.JSON200.PaymentDueDate, "sentToHaciendaAt": parsed.JSON200.SentToHaciendaAt,
		"acceptedByHaciendaAt": parsed.JSON200.AcceptedByHaciendaAt, "rejectedByHaciendaAt": parsed.JSON200.RejectedByHaciendaAt,
		"pdfGeneratedAt": parsed.JSON200.PdfGeneratedAt, "createdAt": parsed.JSON200.CreatedAt,
		"updatedAt": parsed.JSON200.UpdatedAt,
	} {
		if value == nil || value.Format(time.RFC3339Nano) != timestamp {
			t.Fatalf("%s = %v", name, value)
		}
	}
}

func TestAPIErrorPreservesProviderDetails(t *testing.T) {
	err := ErrorFromResponse(&http.Response{StatusCode: http.StatusUnprocessableEntity}, []byte(`{
		"error":"VALIDATION_FAILED","message":"missing receiver","path":"/documents/invoice",
		"details":"receiverId is required","validationErrors":{"receiverId":"required"},"metadata":{"requestId":"abc"}
	}`))
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("error type = %T", err)
	}
	if apiErr.ErrorCode != "VALIDATION_FAILED" || apiErr.ValidationErrors["receiverId"] != "required" || apiErr.Metadata["requestId"] != "abc" {
		t.Fatalf("unexpected API error: %#v", apiErr)
	}
}

func TestClientHonorsCanceledContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()
	client, err := NewAPIKeyClient(server.URL, "secret")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := client.GetTaxCodes(ctx, nil); err == nil {
		t.Fatal("expected context cancellation error")
	}
}
