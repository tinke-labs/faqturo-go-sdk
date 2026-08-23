package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type endpointContract struct {
	operationID           string
	methodName            string
	path                  string
	httpMethod            string
	status                int
	responseContentType   string
	responseBody          string
	responseHasJSONSchema bool
	requestContentType    string
}

func main() {
	specPath := flag.String("spec", "openapi/faqturo-sdk.codegen.json", "OpenAPI specification")
	outPath := flag.String("out", "endpoint_contract_test.go", "generated test output")
	flag.Parse()

	data, err := os.ReadFile(*specPath)
	if err != nil {
		fail(err)
	}

	var spec map[string]any
	if err := json.Unmarshal(data, &spec); err != nil {
		fail(err)
	}

	components := asMap(asMap(spec["components"])["schemas"])
	contracts := collectContracts(asMap(spec["paths"]), components)
	if len(contracts) == 0 {
		fail(fmt.Errorf("no endpoint contracts found in %s", *specPath))
	}

	if err := os.WriteFile(*outPath, []byte(render(contracts)), 0o644); err != nil {
		fail(err)
	}
}

func collectContracts(paths, components map[string]any) []endpointContract {
	pathNames := sortedKeys(paths)
	contracts := make([]endpointContract, 0)
	for _, pathName := range pathNames {
		pathItem := asMap(paths[pathName])
		for _, httpMethod := range sortedKeys(pathItem) {
			if httpMethod == "parameters" || httpMethod == "$ref" {
				continue
			}
			operation := asMap(pathItem[httpMethod])
			operationID, _ := operation["operationId"].(string)
			if operationID == "" {
				continue
			}

			requestContentType := chooseContentType(asMap(asMap(operation["requestBody"])["content"]))
			responses := asMap(operation["responses"])
			for _, statusName := range sortedKeys(responses) {
				status, err := strconv.Atoi(statusName)
				if err != nil || status < 200 || status >= 300 {
					continue
				}

				response := asMap(responses[statusName])
				content := asMap(response["content"])
				responseContentType := chooseContentType(content)
				responseBody := responseBody(content, responseContentType, components)
				_, hasJSONSchema := responseJSONSchema(content)

			methodName := upperFirst(operationID)
			if requestContentType != "" {
				methodName = upperFirst(operationID) + "WithBody"
			}

				contracts = append(contracts, endpointContract{
					operationID:           operationID,
					methodName:            methodName,
					path:                  samplePath(pathName, pathItem, operation),
					httpMethod:            strings.ToUpper(httpMethod),
					status:                status,
					responseContentType:   responseContentType,
					responseBody:          responseBody,
					responseHasJSONSchema: hasJSONSchema,
					requestContentType:    requestContentType,
				})
			}
		}
	}
	return contracts
}

func responseBody(content map[string]any, contentType string, components map[string]any) string {
	if contentType == "" {
		return ""
	}
	if contentType == "application/xml" {
		return `<?xml version="1.0"?><document/>`
	}

	media := asMap(content[contentType])
	schema := asMap(media["schema"])
	if len(schema) == 0 {
		return ""
	}

	sample := sampleSchema(schema, components, map[string]bool{})
	encoded, err := json.Marshal(sample)
	if err != nil {
		fail(fmt.Errorf("marshal response sample: %w", err))
	}
	return string(encoded)
}

func responseJSONSchema(content map[string]any) (map[string]any, bool) {
	media := asMap(content["application/json"])
	schema := asMap(media["schema"])
	return schema, len(schema) > 0
}

func sampleSchema(schema map[string]any, components map[string]any, seen map[string]bool) any {
	if len(schema) == 0 {
		return nil
	}

	if ref, ok := schema["$ref"].(string); ok {
		name := strings.TrimPrefix(ref, "#/components/schemas/")
		if seen[name] {
			return nil
		}
		seen[name] = true
		value := sampleSchema(asMap(components[name]), components, seen)
		delete(seen, name)
		return value
	}

	for _, keyword := range []string{"oneOf", "anyOf"} {
		if choices, ok := schema[keyword].([]any); ok {
			for _, choice := range choices {
				if value := sampleSchema(asMap(choice), components, seen); value != nil {
					return value
				}
			}
		}
	}

	if allOf, ok := schema["allOf"].([]any); ok {
		merged := map[string]any{}
		var first any
		for _, part := range allOf {
			value := sampleSchema(asMap(part), components, seen)
			if first == nil {
				first = value
			}
			if object, ok := value.(map[string]any); ok {
				for key, field := range object {
					merged[key] = field
				}
			}
		}
		if len(merged) > 0 {
			return merged
		}
		return first
	}

	if enum, ok := schema["enum"].([]any); ok && len(enum) > 0 {
		return enum[0]
	}

	typeName, _ := schema["type"].(string)
	switch typeName {
	case "object":
		object := map[string]any{}
		properties := asMap(schema["properties"])
		for _, name := range sortedKeys(properties) {
			object[name] = sampleSchema(asMap(properties[name]), components, seen)
		}
		if len(object) == 0 {
			switch additional := schema["additionalProperties"].(type) {
			case map[string]any:
				object["value"] = sampleSchema(additional, components, seen)
			case bool:
				if additional {
					object["value"] = "contract"
				}
			}
		}
		return object
	case "array":
		itemSchema := asMap(schema["items"])
		if len(itemSchema) == 0 {
			return []any{}
		}
		return []any{sampleSchema(itemSchema, components, seen)}
	case "string":
		format, _ := schema["format"].(string)
		switch format {
		case "date-time":
			return "2026-08-11T10:16:58.882370375-06:00"
		case "date":
			return "2026-08-11"
		default:
			return "contract"
		}
	case "integer":
		return 1
	case "number":
		return 1.25
	case "boolean":
		return true
	}

	if example, ok := schema["example"]; ok {
		return example
	}
	return map[string]any{}
}

func samplePath(path string, pathItem, operation map[string]any) string {
	parameters := map[string]string{}
	for _, source := range []map[string]any{pathItem, operation} {
		for _, raw := range asSlice(source["parameters"]) {
			parameter := asMap(raw)
			if parameter["in"] != "path" {
				continue
			}
			name, _ := parameter["name"].(string)
			schema := asMap(parameter["schema"])
			value := "test"
			if schema["type"] == "integer" || schema["type"] == "number" {
				value = "1"
			}
			parameters[name] = value
		}
	}
	for name, value := range parameters {
		path = strings.ReplaceAll(path, "{"+name+"}", value)
	}
	return path
}

func render(contracts []endpointContract) string {
	var out strings.Builder
	out.WriteString(`// Code generated by generate-endpoint-contracts; DO NOT EDIT.
package faqturo

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

type endpointResponseContract struct {
	operationID           string
	methodName            string
	path                  string
	httpMethod            string
	status                int
	responseContentType   string
	responseBody          string
	responseHasJSONSchema bool
	requestContentType    string
}

const endpointContractTimestamp = "2026-08-11T10:16:58.882370375-06:00"

var endpointResponseContracts = []endpointResponseContract{
`)
	for _, contract := range contracts {
		fmt.Fprintf(&out, "\t{operationID: %q, methodName: %q, path: %q, httpMethod: %q, status: %d, responseContentType: %q, responseBody: %q, responseHasJSONSchema: %t, requestContentType: %q},\n", contract.operationID, contract.methodName, contract.path, contract.httpMethod, contract.status, contract.responseContentType, contract.responseBody, contract.responseHasJSONSchema, contract.requestContentType)
	}
	out.WriteString(`}

func TestEveryEndpointResponseContract(t *testing.T) {
	for _, contract := range endpointResponseContracts {
		contract := contract
		t.Run(fmt.Sprintf("%s_%d", contract.operationID, contract.status), func(t *testing.T) {
			requests := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests++
				if r.Method != contract.httpMethod {
					t.Errorf("method = %s, want %s", r.Method, contract.httpMethod)
				}
				if r.URL.Path != "/api"+contract.path {
					t.Errorf("path = %s, want /api%s", r.URL.Path, contract.path)
				}
				if got := r.Header.Get("X-API-KEY"); got != "contract-key" {
					t.Errorf("X-API-KEY = %q", got)
				}
				if got := r.Header.Get("API-Version"); got != "v1" {
					t.Errorf("API-Version = %q", got)
				}
				if contract.responseContentType != "" {
					w.Header().Set("Content-Type", contract.responseContentType)
				}
				w.WriteHeader(contract.status)
				_, _ = io.WriteString(w, contract.responseBody)
			}))
			defer server.Close()

			client, err := NewAPIKeyClient(server.URL, "contract-key", WithAPIVersion("v1"))
			if err != nil {
				t.Fatal(err)
			}
			method := reflect.ValueOf(client).MethodByName(contract.methodName)
			if !method.IsValid() {
				t.Fatalf("generated client method %s does not exist", contract.methodName)
			}

			results := method.Call(endpointContractArguments(method.Type(), contract.requestContentType))
			if len(results) != 2 {
				t.Fatalf("method returned %d values, want 2", len(results))
			}
			if !results[1].IsNil() {
				t.Fatalf("endpoint returned error: %v", results[1].Interface())
			}
			if results[0].IsNil() {
				t.Fatal("endpoint returned a nil response")
			}

			response := results[0].Elem()
			httpResponse := response.FieldByName("HTTPResponse")
			if !httpResponse.IsValid() || httpResponse.IsNil() {
				t.Fatal("response did not include HTTPResponse")
			}
			actualHTTPResponse := httpResponse.Interface().(*http.Response)
			if actualHTTPResponse.StatusCode != contract.status {
				t.Fatalf("status = %d, want %d", actualHTTPResponse.StatusCode, contract.status)
			}

			body := response.FieldByName("Body")
			if !body.IsValid() || string(body.Bytes()) != contract.responseBody {
				t.Fatalf("raw response body = %q, want %q", body.Bytes(), contract.responseBody)
			}

			jsonPayload := response.FieldByName(fmt.Sprintf("JSON%d", contract.status))
			if contract.responseHasJSONSchema {
				if !jsonPayload.IsValid() {
					t.Fatalf("response has no JSON%d field", contract.status)
				}
				if jsonPayload.Kind() == reflect.Pointer && jsonPayload.IsNil() {
					t.Fatalf("JSON%d payload is nil", contract.status)
				}
				assertEndpointContractTimestamps(t, jsonPayload, "JSON"+fmt.Sprint(contract.status))
			}
			if requests != 1 {
				t.Fatalf("server received %d requests, want 1", requests)
			}
		})
	}
}

var endpointContractReaderType = reflect.TypeOf((*io.Reader)(nil)).Elem()

func endpointContractArguments(methodType reflect.Type, requestContentType string) []reflect.Value {
	argumentCount := methodType.NumIn()
	if methodType.IsVariadic() {
		argumentCount--
	}
	arguments := make([]reflect.Value, 0, argumentCount)
	for index := 0; index < argumentCount; index++ {
		parameterType := methodType.In(index)
		switch {
		case index == 0:
			arguments = append(arguments, reflect.ValueOf(context.Background()))
		case parameterType == endpointContractReaderType:
			arguments = append(arguments, reflect.ValueOf(strings.NewReader("{}")))
		case parameterType.Kind() == reflect.String:
			if index+1 < argumentCount && methodType.In(index+1) == endpointContractReaderType {
				arguments = append(arguments, reflect.ValueOf(requestContentType))
			} else {
				value := reflect.New(parameterType).Elem()
				value.SetString("test")
				arguments = append(arguments, value)
			}
		case parameterType.Kind() >= reflect.Int && parameterType.Kind() <= reflect.Int64:
			value := reflect.New(parameterType).Elem()
			value.SetInt(1)
			arguments = append(arguments, value)
		case parameterType.Kind() >= reflect.Uint && parameterType.Kind() <= reflect.Uint64:
			value := reflect.New(parameterType).Elem()
			value.SetUint(1)
			arguments = append(arguments, value)
		default:
			arguments = append(arguments, reflect.Zero(parameterType))
		}
	}
	return arguments
}

func assertEndpointContractTimestamps(t *testing.T, value reflect.Value, path string) {
	if !value.IsValid() {
		return
	}
	if value.Kind() == reflect.Struct && value.Type().Name() == "Date" {
		return
	}
	if value.Type() == reflect.TypeOf(time.Time{}) {
		actual := value.Interface().(time.Time).Format(time.RFC3339Nano)
		if actual != endpointContractTimestamp {
			t.Errorf("%s timestamp = %q, want %q", path, actual, endpointContractTimestamp)
		}
		return
	}

	switch value.Kind() {
	case reflect.Pointer, reflect.Interface:
		if !value.IsNil() {
			assertEndpointContractTimestamps(t, value.Elem(), path)
		}
	case reflect.Slice, reflect.Array:
		for index := 0; index < value.Len(); index++ {
			assertEndpointContractTimestamps(t, value.Index(index), fmt.Sprintf("%s[%d]", path, index))
		}
	case reflect.Map:
		for _, key := range value.MapKeys() {
			assertEndpointContractTimestamps(t, value.MapIndex(key), fmt.Sprintf("%s[%v]", path, key.Interface()))
		}
	case reflect.Struct:
		for index := 0; index < value.NumField(); index++ {
			field := value.Field(index)
			if field.CanInterface() {
				assertEndpointContractTimestamps(t, field, path+"."+value.Type().Field(index).Name)
			}
		}
	}
}
`)
	return out.String()
}

func chooseContentType(content map[string]any) string {
	if _, ok := content["application/json"]; ok {
		return "application/json"
	}
	keys := sortedKeys(content)
	if len(keys) == 0 {
		return ""
	}
	return keys[0]
}

func asMap(value any) map[string]any {
	result, _ := value.(map[string]any)
	return result
}

func asSlice(value any) []any {
	result, _ := value.([]any)
	return result
}

func sortedKeys(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func upperFirst(value string) string {
	if value == "" {
		return value
	}
	return strings.ToUpper(value[:1]) + value[1:]
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
