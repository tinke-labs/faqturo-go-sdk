package faqturo

import (
	"bytes"
	"encoding/json"
	"errors"
	"math/big"
)

// Decimal is an exact base-10 JSON number. It never converts through float64.
type Decimal struct{ value string }

func NewDecimal(value string) (Decimal, error) {
	d := Decimal{}
	if err := d.UnmarshalJSON([]byte(value)); err != nil {
		return Decimal{}, err
	}
	return d, nil
}

func MustDecimal(value string) Decimal {
	d, err := NewDecimal(value)
	if err != nil {
		panic(err)
	}
	return d
}

func (d Decimal) String() string { return d.value }

func (d Decimal) MarshalJSON() ([]byte, error) {
	if d.value == "" {
		return []byte("0"), nil
	}
	if !validDecimal(d.value) {
		return nil, errors.New("faqturo: invalid decimal")
	}
	return []byte(d.value), nil
}

func (d *Decimal) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if bytes.Equal(data, []byte("null")) {
		d.value = ""
		return nil
	}
	if len(data) == 0 || data[0] == '"' || !validDecimal(string(data)) {
		return errors.New("faqturo: decimal must be a JSON number")
	}
	d.value = string(data)
	return nil
}

func validDecimal(value string) bool {
	n := json.Number(value)
	_, ok := new(big.Rat).SetString(n.String())
	return ok
}
