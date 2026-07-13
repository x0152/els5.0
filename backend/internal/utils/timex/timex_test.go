package timex

import (
	"testing"
	"time"
)

func TestMSKLoaded(t *testing.T) {
	if MSK == nil {
		t.Fatal("MSK is nil")
	}
}

func TestTimeLocalIsMSK(t *testing.T) {
	if time.Local != MSK {
		t.Fatalf("time.Local is not MSK: %v", time.Local)
	}
}

func TestNowInMSK(t *testing.T) {
	n := Now()
	if n.Location() != MSK {
		t.Fatalf("Now() not in MSK: %v", n.Location())
	}
}

func TestSystemClockInMSK(t *testing.T) {
	n := System().Now()
	if n.Location() != MSK {
		t.Fatalf("System().Now() not in MSK: %v", n.Location())
	}
}

func TestFrozenClockNormalizesToMSK(t *testing.T) {
	utc := time.Date(2026, 4, 17, 9, 0, 0, 0, time.UTC)
	fc := NewFrozen(utc)
	if fc.Now().Location() != MSK {
		t.Fatalf("frozen clock not in MSK: %v", fc.Now().Location())
	}
	fc.Set(time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC))
	if fc.Now().Location() != MSK {
		t.Fatalf("after Set: not in MSK: %v", fc.Now().Location())
	}
}

func TestDateConstructorMSK(t *testing.T) {
	d := Date(2026, time.April, 17, 12, 0, 0, 0)
	if d.Location() != MSK {
		t.Fatalf("Date not in MSK: %v", d.Location())
	}
	if d.Hour() != 12 {
		t.Fatalf("hour lost: %d", d.Hour())
	}
}

func TestParseRFC3339ReturnsMSK(t *testing.T) {
	cases := []string{
		"2026-04-17T12:00:00+03:00",
		"2026-04-17T09:00:00Z",
	}
	for _, s := range cases {
		got, err := ParseRFC3339(s)
		if err != nil {
			t.Fatalf("parse %q: %v", s, err)
		}
		if got.Location() != MSK {
			t.Fatalf("parse %q: not in MSK: %v", s, got.Location())
		}
	}
}

func TestParseDateMSK(t *testing.T) {
	got, err := ParseDate("2026-04-17")
	if err != nil {
		t.Fatal(err)
	}
	if got.Location() != MSK {
		t.Fatalf("not in MSK: %v", got.Location())
	}
	if got.Hour() != 0 || got.Minute() != 0 {
		t.Fatalf("expected 00:00, got %02d:%02d", got.Hour(), got.Minute())
	}
}

func TestEndOfDayDSTSafe(t *testing.T) {
	base := Date(2026, time.April, 17, 15, 30, 45, 0)
	end := EndOfDay(base)
	wantY, wantM, wantD := 2026, time.April, 17
	y, m, d := end.Date()
	if y != wantY || m != wantM || d != wantD {
		t.Fatalf("EndOfDay date shifted: %d-%02d-%02d", y, m, d)
	}
	if end.Hour() != 23 || end.Minute() != 59 || end.Second() != 59 {
		t.Fatalf("EndOfDay time wrong: %02d:%02d:%02d", end.Hour(), end.Minute(), end.Second())
	}
}

func TestDaysBetween(t *testing.T) {
	a := Date(2026, time.April, 1, 23, 0, 0, 0)
	b := Date(2026, time.April, 10, 1, 0, 0, 0)
	if got := DaysBetween(a, b); got != 9 {
		t.Fatalf("expected 9, got %d", got)
	}
}

func TestFormatRFC3339OffsetMSK(t *testing.T) {
	t0 := Date(2026, time.April, 17, 12, 0, 0, 0)
	s := FormatRFC3339(t0)
	if s != "2026-04-17T12:00:00+03:00" {
		t.Fatalf("unexpected format: %s", s)
	}
}

func TestResolveTimezoneNamePrefersTZ(t *testing.T) {
	t.Setenv("TZ", "Europe/Berlin")
	t.Setenv("POSTGRES_TIMEZONE", "Europe/Moscow")

	got := resolveTimezoneName()
	if got != "Europe/Berlin" {
		t.Fatalf("expected Europe/Berlin, got %s", got)
	}
}

func TestResolveTimezoneNameUsesPostgresTimezone(t *testing.T) {
	t.Setenv("TZ", "")
	t.Setenv("POSTGRES_TIMEZONE", "Europe/Paris")

	got := resolveTimezoneName()
	if got != "Europe/Paris" {
		t.Fatalf("expected Europe/Paris, got %s", got)
	}
}

func TestResolveTimezoneNameDefault(t *testing.T) {
	t.Setenv("TZ", "")
	t.Setenv("POSTGRES_TIMEZONE", "")

	got := resolveTimezoneName()
	if got != defaultTZ {
		t.Fatalf("expected %s, got %s", defaultTZ, got)
	}
}
