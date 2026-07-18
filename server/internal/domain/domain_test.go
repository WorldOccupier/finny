package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNewDecimalAcceptsValidNonNegativeValues(t *testing.T) {
	for _, input := range []string{"0", "12.50", "12345678901234567890.123456789"} {
		value, err := NewDecimal(input)
		if err != nil {
			t.Fatalf("NewDecimal(%q): %v", input, err)
		}
		if value.String() == "" {
			t.Errorf("NewDecimal(%q) produced an empty canonical value", input)
		}
	}
}

func TestDecimalRejectsInvalidValues(t *testing.T) {
	for _, input := range []string{"", "not-a-number", "-1"} {
		if _, err := NewDecimal(input); err == nil {
			t.Errorf("NewDecimal(%q) succeeded", input)
		}
	}

	for _, input := range []string{"12.50", "null", "-1"} {
		var value Decimal
		if err := json.Unmarshal([]byte(input), &value); err == nil {
			t.Errorf("json.Unmarshal(%s) succeeded", input)
		}
	}
}

func TestDecimalJSONUsesStrings(t *testing.T) {
	value, err := NewDecimal("12.50")
	if err != nil {
		t.Fatal(err)
	}

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `"12.5"` {
		t.Fatalf("JSON = %s", data)
	}

	var roundTripped Decimal
	if err := json.Unmarshal(data, &roundTripped); err != nil {
		t.Fatal(err)
	}
	if !roundTripped.Equal(value.Decimal) {
		t.Fatalf("round trip = %s", roundTripped.String())
	}
}

func TestValueTypesMapToCurrencies(t *testing.T) {
	ukCurrency, err := UK_GBP.Currency()
	if err != nil || !UK_GBP.Valid() || ukCurrency != CURRENCY_GBP {
		t.Errorf("UK_GBP mapping is invalid")
	}
	indiaCurrency, err := INDIA_INR.Currency()
	if err != nil || !INDIA_INR.Valid() || indiaCurrency != CURRENCY_INR {
		t.Errorf("INDIA_INR mapping is invalid")
	}
	if ValueType("OTHER").Valid() {
		t.Errorf("unsupported value type is valid")
	}
	if _, err := ValueType("OTHER").Currency(); err == nil {
		t.Errorf("unsupported value type returned a currency")
	}
}

func TestDomainValuesValidate(t *testing.T) {
	value, err := NewDecimal("100.00")
	if err != nil {
		t.Fatal(err)
	}
	asset := Asset{ID: 0, Name: "Savings", Values: []AssetValue{
		{Type: UK_GBP, Value: value},
		{Type: INDIA_INR, Value: value},
	}}
	if err := asset.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (Asset{Name: "  "}).Validate(); err == nil {
		t.Fatal("empty asset name was accepted")
	}
	if err := (SpendingLimit{Key: "rent", Amount: value, Currency: CURRENCY_GBP}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (SpendingLimit{Key: "rent", Amount: value, Currency: Currency("USD")}).Validate(); err == nil {
		t.Fatal("unsupported spending-limit currency was accepted")
	}
}

func TestLondonTimestampParsingAndFormatting(t *testing.T) {
	parsed, err := ParseLondonTimestamp("2026-01-15T12:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	if got := FormatLondonTimestamp(parsed); got != "2026-01-15T12:00:00Z" {
		t.Fatalf("winter timestamp = %q", got)
	}

	dst, err := ParseLondonTimestamp("2026-07-15T12:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	if got := FormatLondonTimestamp(dst); got != "2026-07-15T13:00:00+01:00" {
		t.Fatalf("summer timestamp = %q", got)
	}
	if dst.Location() != LondonLocation() {
		t.Fatalf("timestamp location = %v", dst.Location())
	}
}

func TestDashboardJSONShape(t *testing.T) {
	value, err := NewDecimal("1000.00")
	if err != nil {
		t.Fatal(err)
	}
	dashboard := Dashboard{
		Assets:        []Asset{{ID: 0, Name: "Savings", Values: []AssetValue{{Type: UK_GBP, Value: value}}}},
		CurrentFXRate: value,
		CurrentTotals: DashboardTotals{Country: []TotalValue{{Type: UK_GBP, Value: value}}, Combined: []CombinedTotal{{Currency: CURRENCY_GBP, Value: value}}},
		Income:        IncomeTotals{UserOneGBP: value, UserTwoGBP: value},
	}
	data, err := json.Marshal(dashboard)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{`"type":"UKGBP"`, `"value":"1000"`, `"currency":"GBP"`} {
		if !strings.Contains(string(data), expected) {
			t.Errorf("JSON %s missing %s", data, expected)
		}
	}
}
