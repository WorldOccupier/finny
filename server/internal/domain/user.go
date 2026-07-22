package domain

import "fmt"

type UserID string

const (
	USER_ONE UserID = "user_one"
	USER_TWO UserID = "user_two"
)

type User struct {
	ID UserID `json:"id"`
}

func (u User) Validate() error {
	if u.ID != USER_ONE && u.ID != USER_TWO {
		return fmt.Errorf("invalid user %q", u.ID)
	}
	return nil
}
