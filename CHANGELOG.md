# Changelog

## v1.1.0 - 2026-10-16

- Add document type and payment methods to fiscal preview requests, and expose the calculated `vatRefunded` amount.
- Document temporary Hacienda throttling responses with HTTP 503, `TAX_AUTHORITY_RATE_LIMIT_ERROR`, and `Retry-After`.
- Refresh the generated Go client, OpenAPI contract, and endpoint response tests from backend `15a5990`.
- Align CABYS search and fiscal contact validation constraints with Hacienda's published formats.
