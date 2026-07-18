package domain

import "fmt"

type ValueType string

const (
	UK_GBP    ValueType = "UKGBP"
	INDIA_INR ValueType = "INDIAINR"
)

func (t ValueType) Valid() bool {
	return t == UK_GBP || t == INDIA_INR
}

func (t ValueType) Currency() (Currency, error) {
	switch t {
	case UK_GBP:
		return CURRENCY_GBP, nil
	case INDIA_INR:
		return CURRENCY_INR, nil
	default:
		return "", fmt.Errorf("invalid value type %q", t)
	}
}
