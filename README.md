# Faqturo Go SDK

Official Go client for the API-key integration surface of Faqturo. Requires Go 1.25 or newer.

```bash
go get github.com/tinke-labs/faqturo-go-sdk@v0.1.0
```

## Client

```go
client, err := faqturo.NewAPIKeyClient(
    "https://api.faqturo.com",
    apiKey,
    faqturo.WithAPIVersion("v1"),
)
```

`NewAPIKeyClient` sends `X-API-KEY` and `API-Version: v1` automatically. Use
`WithHTTPClient`, `WithTimeout`, and `WithRequestEditor` to configure TLS certificates,
timeouts, proxies, tracing, or other transport behavior.

## Common workflows

- Invoicing: construct an `InvoiceRequest` and call `CreateInvoiceWithResponse`.
- Queries: use the typed `GetDocumentsWithResponse`, catalog, client, issuer, and tax-authority methods.
- Files: PDF/XML methods return typed responses while CSV export responses preserve their binary body.
- Multipart: generated helpers cover receiver XML, fiscal certificates, and tenant logos.
- Errors: call `ErrorFromResponse(response, body)` for a structured `*APIError` containing validation errors and metadata.

Money and rates use `Decimal`, which preserves the exact JSON number and never passes through `float64`:

```go
amount := faqturo.MustDecimal("1234567890.123456789")
```

## Generation

The source contract is [`openapi/faqturo-sdk.json`](openapi/faqturo-sdk.json). Generated
code is committed. Regenerate deterministically with:

```bash
go generate ./...
go test ./...
go vet ./...
```

`openapi/decimal-overlay.yaml` maps every `number/decimal` schema to `faqturo.Decimal`.
The generated `openapi/faqturo-sdk.codegen.json` is retained to make the overlay result auditable.

## Branding

Faqturo and Tinke Labs names and marks belong to Tinke Labs. The Apache-2.0 license
covers this SDK's source code; it does not grant trademark rights or imply endorsement
of modified or third-party distributions.
