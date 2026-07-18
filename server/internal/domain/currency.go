package domain

type Currency string

const (
	CURRENCY_GBP Currency = "GBP"
	CURRENCY_INR Currency = "INR"
)

func (c Currency) Valid() bool {
	return c == CURRENCY_GBP || c == CURRENCY_INR
}
