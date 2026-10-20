package schedule

import "time"

const searchYears = 5

// Next returns the first activation strictly after t, or the zero time if the
// schedule never fires (e.g. "0 0 30 2 *").
func (s *Schedule) Next(t time.Time) time.Time {
	t = t.Truncate(time.Minute).Add(time.Minute)
	limit := t.Year() + searchYears
	loc := t.Location()

	for t.Year() <= limit {
		if !has(s.Month, int(t.Month())) {
			t = advance(t, time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, loc))
			continue
		}
		if !s.dayMatches(t) {
			t = advance(t, time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, 0, loc))
			continue
		}
		if !has(s.Hour, t.Hour()) {
			t = t.Add(time.Duration(60-t.Minute()) * time.Minute)
			continue
		}
		if !has(s.Minute, t.Minute()) {
			t = t.Add(time.Minute)
			continue
		}
		return t
	}
	return time.Time{}
}

// Upcoming returns up to n activations after t.
func (s *Schedule) Upcoming(t time.Time, n int) []time.Time {
	var out []time.Time
	for len(out) < n {
		t = s.Next(t)
		if t.IsZero() {
			break
		}
		out = append(out, t)
	}
	return out
}

// Matches reports whether the schedule fires during the minute containing t.
func (s *Schedule) Matches(t time.Time) bool {
	return has(s.Minute, t.Minute()) && has(s.Hour, t.Hour()) &&
		has(s.Month, int(t.Month())) && s.dayMatches(t)
}

// When both day fields are restricted, cron fires if either one matches.
func (s *Schedule) dayMatches(t time.Time) bool {
	dom := has(s.Dom, t.Day())
	dow := has(s.Dow, int(t.Weekday()))
	if s.domStar || s.dowStar {
		return dom && dow
	}
	return dom || dow
}

// time.Date can land before t when a DST shift happens at local midnight, so
// fall back to stepping to the next hour boundary.
func advance(t, next time.Time) time.Time {
	if next.After(t) {
		return next
	}
	return t.Add(time.Duration(60-t.Minute()) * time.Minute)
}

func has(bits uint64, v int) bool {
	return bits&(1<<uint(v)) != 0
}
