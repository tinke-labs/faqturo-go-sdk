# Changelog

## v2.1.0 - 2026-11-25

- Add `GetClientAccountStatementXlsx` for `GET /accounts-receivable/clients/{clientId}/statement.xlsx` (invoices with settlement date and abonos).
- Add `GetReceivablesReportXlsx` and `GetReceivablesReportPdf` for `GET /accounts-receivable/report.xlsx|.pdf`, filterable by `client_id`, `from`, `to` and `date_field`.
- Add the optional `Idempotency-Key` parameter to `CreatePaymentParams`; a retry with the same request returns the existing payment.
- Refresh the generated Go client, OpenAPI contract and endpoint response tests from backend `65e12b7`.

## v2.0.0 - unreleased

- Rename receiver message state fields and multipart parameter to `receiverMessageDecision`.
- Move the Go module to `github.com/tinke-labs/faqturo-go-sdk/v2` for this breaking contract change.

## v1.1.0 - 2026-10-16

- Add document type and payment methods to fiscal preview requests, and expose the calculated `vatRefunded` amount.
- Document temporary Hacienda throttling responses with HTTP 503, `TAX_AUTHORITY_RATE_LIMIT_ERROR`, and `Retry-After`.
- Refresh the generated Go client, OpenAPI contract, and endpoint response tests from backend `15a5990`.
- Align CABYS search and fiscal contact validation constraints with Hacienda's published formats.
