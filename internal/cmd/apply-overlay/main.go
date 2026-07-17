// Command apply-overlay applies the repository's deliberately narrow decimal overlay.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

func main() {
	in := flag.String("in", "", "input OpenAPI document")
	out := flag.String("out", "", "output OpenAPI document")
	overlay := flag.String("overlay", "", "overlay document (required and versioned for auditability)")
	flag.Parse()
	if *in == "" || *out == "" || *overlay == "" {
		fmt.Fprintln(os.Stderr, "-in, -out and -overlay are required")
		os.Exit(2)
	}
	if _, err := os.Stat(*overlay); err != nil {
		fatal(err)
	}
	b, err := os.ReadFile(*in)
	if err != nil {
		fatal(err)
	}
	var document any
	if err := json.Unmarshal(b, &document); err != nil {
		fatal(err)
	}
	// oapi-codegen v2.7.0 targets OpenAPI 3.0. The source remains the verbatim 3.1
	// backend export; only this deterministic code-generation view is downgraded.
	if root, ok := document.(map[string]any); ok {
		root["openapi"] = "3.0.3"
	}
	apply(document)
	b, err = json.MarshalIndent(document, "", "  ")
	if err != nil {
		fatal(err)
	}
	b = append(b, '\n')
	if err := os.WriteFile(*out, b, 0o644); err != nil {
		fatal(err)
	}
}

func apply(value any) {
	switch value := value.(type) {
	case map[string]any:
		if value["type"] == "number" && value["format"] == "decimal" {
			value["x-go-type"] = "Decimal"
		}
		for _, child := range value {
			apply(child)
		}
	case []any:
		for _, child := range value {
			apply(child)
		}
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
