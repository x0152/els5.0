package timex

import (
	"log"
	"os"
	"strings"
	"time"
)

const (
	DateFormat     = "2006-01-02"
	DateTimeFormat = time.RFC3339
	defaultTZ      = "Europe/Moscow"
)

var MSK *time.Location

func init() {
	tzName := resolveTimezoneName()
	loc, err := time.LoadLocation(tzName)
	if err != nil && tzName != defaultTZ {
		log.Printf("[timex] timezone %q unavailable, fallback to %q: %v", tzName, defaultTZ, err)
		loc, err = time.LoadLocation(defaultTZ)
	}
	if err != nil {
		log.Printf("[timex] %s tzdata unavailable, fallback to fixed UTC+3: %v", defaultTZ, err)
		loc = time.FixedZone("MSK", 3*60*60)
	}
	MSK = loc
	time.Local = loc
}

func resolveTimezoneName() string {
	if tz := strings.TrimSpace(os.Getenv("TZ")); tz != "" {
		return tz
	}
	if tz := strings.TrimSpace(os.Getenv("POSTGRES_TIMEZONE")); tz != "" {
		return tz
	}
	return defaultTZ
}

func Now() time.Time {
	return System().Now()
}

func Date(year int, month time.Month, day, hour, min, sec, nsec int) time.Time {
	return time.Date(year, month, day, hour, min, sec, nsec, MSK)
}

func FormatDate(t time.Time) string {
	return t.In(MSK).Format(DateFormat)
}

func ParseDate(s string) (time.Time, error) {
	return time.ParseInLocation(DateFormat, s, MSK)
}

func FormatRFC3339(t time.Time) string {
	return t.In(MSK).Format(DateTimeFormat)
}

func ParseRFC3339(s string) (time.Time, error) {
	t, err := time.Parse(DateTimeFormat, s)
	if err != nil {
		return time.Time{}, err
	}
	return t.In(MSK), nil
}

func StartOfDay(t time.Time) time.Time {
	y, m, d := t.In(MSK).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, MSK)
}

func EndOfDay(t time.Time) time.Time {
	return StartOfDay(t).AddDate(0, 0, 1).Add(-time.Nanosecond)
}

func StartOfMonth(t time.Time) time.Time {
	y, m, _ := t.In(MSK).Date()
	return time.Date(y, m, 1, 0, 0, 0, 0, MSK)
}

func DaysBetween(from, to time.Time) int {
	ya, ma, da := from.In(MSK).Date()
	yb, mb, db := to.In(MSK).Date()
	a := time.Date(ya, ma, da, 0, 0, 0, 0, time.UTC)
	b := time.Date(yb, mb, db, 0, 0, 0, 0, time.UTC)
	return int(b.Sub(a) / (24 * time.Hour))
}
