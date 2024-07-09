package util

import "time"

func IsTimeUnitValid(unit time.Duration) bool {
	switch unit {
	case time.Nanosecond, time.Microsecond, time.Millisecond, time.Second, time.Minute, time.Hour:
		return true
	default:
		return false
	}
}
