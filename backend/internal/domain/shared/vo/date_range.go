package vo

import "time"

func DateRangeWithin(childStart, childEnd, parentStart, parentEnd time.Time) bool {
	start := childStart.UTC()
	end := childEnd.UTC()
	return !start.Before(parentStart.UTC()) && !end.After(parentEnd.UTC())
}
