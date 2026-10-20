package schedule

import (
	"fmt"
	"math/bits"
	"strings"
)

var monthNames = []string{"", "January", "February", "March", "April", "May", "June",
	"July", "August", "September", "October", "November", "December"}

var dayNames = []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}

// Describe renders the schedule as an English sentence.
func (s *Schedule) Describe() string {
	var b strings.Builder
	b.WriteString(s.describeTime())

	day := s.describeDays()
	if day != "" {
		b.WriteString(", ")
		b.WriteString(day)
	}
	if s.raw[3] != "*" {
		b.WriteString(", in ")
		b.WriteString(joinValues(s.Month, 1, 12, func(v int) string { return monthNames[v] }))
	}
	return b.String()
}

func (s *Schedule) describeTime() string {
	minStar := s.raw[0] == "*"
	hourStar := s.raw[1] == "*"
	minCount := bits.OnesCount64(s.Minute)
	hourCount := bits.OnesCount64(s.Hour)

	switch {
	case minStar && hourStar:
		return "Every minute"
	case minCount == 1 && hourCount == 1:
		return "At " + clock(first(s.Hour), first(s.Minute))
	case minCount == 1 && hourCount <= 4:
		var times []string
		for h := 0; h < 24; h++ {
			if has(s.Hour, h) {
				times = append(times, clock(h, first(s.Minute)))
			}
		}
		return "At " + joinList(times)
	}

	var minPart string
	if step, ok := stepOf(s.raw[0]); ok {
		minPart = fmt.Sprintf("Every %d minutes", step)
	} else if minStar {
		minPart = "Every minute"
	} else if minCount == 1 {
		minPart = fmt.Sprintf("At minute %d", first(s.Minute))
	} else {
		minPart = "At minutes " + joinValues(s.Minute, 0, 59, itoa)
	}

	switch {
	case hourStar:
		if minCount == 1 {
			return minPart + " past every hour"
		}
		return minPart
	case hourCount == 1:
		h := first(s.Hour)
		return fmt.Sprintf("%s between %s and %s", minPart, clock(h, 0), clock(h, 59))
	}
	if step, ok := stepOf(s.raw[1]); ok {
		return fmt.Sprintf("%s, every %d hours", minPart, step)
	}
	return minPart + ", during hours " + joinValues(s.Hour, 0, 23, func(v int) string {
		return fmt.Sprintf("%02d", v)
	})
}

func (s *Schedule) describeDays() string {
	var dom, dow string
	if !s.domStar {
		dom = "on day " + joinValues(s.Dom, 1, 31, itoa) + " of the month"
		if step, ok := stepOf(s.raw[2]); ok {
			dom = fmt.Sprintf("every %d days", step)
		}
	}
	if !s.dowStar {
		dow = "on " + joinValues(s.Dow, 0, 6, func(v int) string { return dayNames[v] })
		if s.Dow == 0b0111110 {
			dow = "on weekdays"
		} else if s.Dow == 0b1000001 {
			dow = "on weekends"
		}
	}
	switch {
	case dom != "" && dow != "":
		return dom + " or " + dow
	case dom != "":
		return dom
	default:
		return dow
	}
}

// joinValues collapses consecutive runs into ranges: 1,2,3,5 -> "1-3 and 5".
func joinValues(set uint64, lo, hi int, name func(int) string) string {
	var parts []string
	for v := lo; v <= hi; v++ {
		if !has(set, v) {
			continue
		}
		end := v
		for end+1 <= hi && has(set, end+1) {
			end++
		}
		switch {
		case end-v >= 2:
			parts = append(parts, name(v)+" through "+name(end))
		case end > v:
			parts = append(parts, name(v), name(end))
		default:
			parts = append(parts, name(v))
		}
		v = end
	}
	return joinList(parts)
}

func joinList(items []string) string {
	switch len(items) {
	case 0:
		return ""
	case 1:
		return items[0]
	case 2:
		return items[0] + " and " + items[1]
	}
	return strings.Join(items[:len(items)-1], ", ") + " and " + items[len(items)-1]
}

func stepOf(raw string) (int, bool) {
	var step int
	if _, err := fmt.Sscanf(raw, "*/%d", &step); err != nil || strings.Contains(raw, ",") {
		return 0, false
	}
	return step, true
}

func first(set uint64) int {
	return bits.TrailingZeros64(set)
}

func clock(h, m int) string {
	return fmt.Sprintf("%02d:%02d", h, m)
}

func itoa(v int) string {
	return fmt.Sprint(v)
}
