package faqturo

import (
	"encoding/json"
	"testing"
)

func TestDecimalPreservesJSONPrecision(t *testing.T) {
	const raw = "12345678901234567890.12345678901234567890"
	d := MustDecimal(raw)
	encoded, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded) != raw {
		t.Fatalf("got %s, want %s", encoded, raw)
	}
	var decoded Decimal
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.String() != raw {
		t.Fatalf("got %s, want %s", decoded.String(), raw)
	}
}

func TestDecimalRejectsJSONString(t *testing.T) {
	var d Decimal
	if json.Unmarshal([]byte(`"1.25"`), &d) == nil {
		t.Fatal("expected error")
	}
}
