package domain

import "time"

func isLondonTime(value time.Time) bool {
	return !value.IsZero() && value.Location() == LondonLocation()
}
