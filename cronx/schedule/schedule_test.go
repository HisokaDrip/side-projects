package schedule

import (
	"testing"
	"time"
)

func at(s string) time.Time {
	t, err := time.ParseInLocation("2006-01-02 15:04", s, time.UTC)
	if err != nil {
		panic(err)
	}
	return t
}

func TestParseErrors(t *testing.T) {
	bad := []string{
		"* * * *",
		"60 * * * *",
		"* 24 * * *",
		"* * 0 * *",
		"* * * 13 *",
		"*/0 * * * *",
		"5-1 * * * *",
		"* * * foo *",
		"1,,2 * * * *",
		"@sometimes",
	}
	for _, expr := range bad {
		if _, err := Parse(expr); err == nil {
			t.Errorf("Parse(%q) succeeded, want error", expr)
		}
	}
}

func TestNext(t *testing.T) {
	tests := []struct {
		expr, from, want string
	}{
		{"* * * * *", "2026-01-01 00:00", "2026-01-01 00:01"},
		{"*/15 * * * *", "2026-01-01 10:07", "2026-01-01 10:15"},
		{"0 9 * * mon-fri", "2026-10-23 10:00", "2026-10-26 09:00"},
		{"30 2 1 * *", "2026-01-31 12:00", "2026-02-01 02:30"},
		{"0 0 29 2 *", "2026-03-01 00:00", "2028-02-29 00:00"},
		{"0 0 * * 7", "2026-10-20 00:00", "2026-10-25 00:00"},
		{"@yearly", "2026-06-15 08:00", "2027-01-01 00:00"},
		{"0 12 13 * fri", "2026-10-20 00:00", "2026-10-23 12:00"},
		{"5/20 * * * *", "2026-01-01 00:30", "2026-01-01 00:45"},
		{"0 0 * dec *", "2026-01-01 00:00", "2026-12-01 00:00"},
	}
	for _, tt := range tests {
		s, err := Parse(tt.expr)
		if err != nil {
			t.Fatalf("Parse(%q): %v", tt.expr, err)
		}
		got := s.Next(at(tt.from))
		if !got.Equal(at(tt.want)) {
			t.Errorf("%q after %s = %s, want %s", tt.expr, tt.from, got.Format("2006-01-02 15:04"), tt.want)
		}
	}
}

func TestNeverFires(t *testing.T) {
	s, _ := Parse("0 0 30 2 *")
	if got := s.Next(at("2026-01-01 00:00")); !got.IsZero() {
		t.Errorf("expected no run, got %s", got)
	}
}

func TestMatches(t *testing.T) {
	s, _ := Parse("*/10 9-17 * * 1-5")
	if !s.Matches(at("2026-10-20 09:30")) {
		t.Error("expected Tuesday 09:30 to match")
	}
	if s.Matches(at("2026-10-24 09:30")) {
		t.Error("expected Saturday not to match")
	}
	if s.Matches(at("2026-10-20 09:35")) {
		t.Error("expected 09:35 not to match")
	}
}

func TestDescribe(t *testing.T) {
	tests := []struct{ expr, want string }{
		{"* * * * *", "Every minute"},
		{"@daily", "At 00:00"},
		{"30 8 * * 1-5", "At 08:30, on weekdays"},
		{"0 9,18 * * sat,sun", "At 09:00 and 18:00, on weekends"},
		{"*/15 * * * *", "Every 15 minutes"},
		{"0 * * * *", "At minute 0 past every hour"},
		{"*/5 9-17 * * *", "Every 5 minutes, during hours 09 through 17"},
		{"0 0 1 * *", "At 00:00, on day 1 of the month"},
		{"0 0 1,15 jan-mar *", "At 00:00, on day 1 and 15 of the month, in January through March"},
		{"15 3 * * mon,wed", "At 03:15, on Monday and Wednesday"},
		{"0 12 13 * fri", "At 12:00, on day 13 of the month or on Friday"},
	}
	for _, tt := range tests {
		s, err := Parse(tt.expr)
		if err != nil {
			t.Fatalf("Parse(%q): %v", tt.expr, err)
		}
		if got := s.Describe(); got != tt.want {
			t.Errorf("Describe(%q) = %q, want %q", tt.expr, got, tt.want)
		}
	}
}

func TestDSTGap(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skip("tzdata not available")
	}
	s, _ := Parse("30 2 * * *")
	runs := s.Upcoming(time.Date(2026, 3, 7, 0, 0, 0, 0, loc), 2)
	want := []time.Time{
		time.Date(2026, 3, 7, 2, 30, 0, 0, loc),
		time.Date(2026, 3, 9, 2, 30, 0, 0, loc),
	}
	if len(runs) != 2 || !runs[0].Equal(want[0]) || !runs[1].Equal(want[1]) {
		t.Errorf("got %v, want %v", runs, want)
	}
}
