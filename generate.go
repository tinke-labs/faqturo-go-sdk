package faqturo

//go:generate go run ./internal/cmd/apply-overlay -overlay openapi/decimal-overlay.yaml -in openapi/faqturo-sdk.json -out openapi/faqturo-sdk.codegen.json
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.7.0 --config oapi-codegen.yaml openapi/faqturo-sdk.codegen.json
//go:generate go run ./internal/cmd/generate-endpoint-contracts -spec openapi/faqturo-sdk.codegen.json -out endpoint_contract_test.go
