package domain

import (
	"fmt"
	"strings"
)

type Account struct {
	ID           string         `json:"id"`
	BankSource   string         `json:"bankSource"`
	AccountLabel string         `json:"accountLabel"`
	Currency     Currency       `json:"currency"`
	Owner        OwnershipScope `json:"owner"`
}

func (a Account) Validate() error {
	if strings.TrimSpace(a.ID) == "" || strings.TrimSpace(a.BankSource) == "" || strings.TrimSpace(a.AccountLabel) == "" {
		return fmt.Errorf("account id, bank source, and account label must not be empty")
	}
	if !a.Currency.Valid() {
		return fmt.Errorf("invalid account currency %q", a.Currency)
	}
	if !validOwner(a.Owner) {
		return fmt.Errorf("invalid account ownership %q", a.Owner)
	}
	return nil
}
