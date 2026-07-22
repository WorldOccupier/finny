package domain

type OwnershipScope string

const (
	OWNERSHIP_USER_ONE OwnershipScope = "user_one"
	OWNERSHIP_USER_TWO OwnershipScope = "user_two"
	OWNERSHIP_JOINT    OwnershipScope = "joint"
)

func VisibleTo(account Account, user UserID) bool {
	if user != USER_ONE && user != USER_TWO {
		return false
	}
	return account.Owner == OWNERSHIP_JOINT || account.Owner == OwnershipScope(user)
}

func validOwner(owner OwnershipScope) bool {
	return owner == OWNERSHIP_USER_ONE || owner == OWNERSHIP_USER_TWO || owner == OWNERSHIP_JOINT
}
