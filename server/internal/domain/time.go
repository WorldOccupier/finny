package domain

import (
	"fmt"
	"time"
)

var londonLocation = func() *time.Location {
	location, err := time.LoadLocation("Europe/London")
	if err != nil {
		panic(fmt.Sprintf("load Europe/London: %v", err))
	}
	return location
}()

func LondonLocation() *time.Location {
	return londonLocation
}

func ParseLondonTimestamp(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timestamp: %w", err)
	}
	return parsed.In(londonLocation), nil
}

func FormatLondonTimestamp(value time.Time) string {
	return value.In(londonLocation).Format(time.RFC3339Nano)
}
