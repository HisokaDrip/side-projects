package schedule

import (
	"fmt"
	"strconv"
	"strings"
)

type field struct {
	name  string
	min   int
	max   int
	names map[string]int
}

var fields = [5]field{
	{name: "minute", min: 0, max: 59},
	{name: "hour", min: 0, max: 23},
	{name: "day of month", min: 1, max: 31},
	{name: "month", min: 1, max: 12, names: map[string]int{
		"jan": 1, "feb": 2, "mar": 3, "apr": 4, "may": 5, "jun": 6,
		"jul": 7, "aug": 8, "sep": 9, "oct": 10, "nov": 11, "dec": 12,
	}},
	{name: "day of week", min: 0, max: 7, names: map[string]int{
		"sun": 0, "mon": 1, "tue": 2, "wed": 3, "thu": 4, "fri": 5, "sat": 6,
	}},
}

var macros = map[string]string{
	"@yearly":   "0 0 1 1 *",
	"@annually": "0 0 1 1 *",
	"@monthly":  "0 0 1 * *",
	"@weekly":   "0 0 * * 0",
	"@daily":    "0 0 * * *",
	"@midnight": "0 0 * * *",
	"@hourly":   "0 * * * *",
}

// Schedule is a parsed five-field cron expression. Each field is a bitset of
// allowed values.
type Schedule struct {
	Expr   string
	Minute uint64
	Hour   uint64
	Dom    uint64
	Month  uint64
	Dow    uint64

	domStar bool
	dowStar bool
	raw     [5]string
}

func Parse(expr string) (*Schedule, error) {
	expr = strings.TrimSpace(expr)
	spec := expr
	if strings.HasPrefix(spec, "@") {
		m, ok := macros[strings.ToLower(spec)]
		if !ok {
			return nil, fmt.Errorf("unknown macro %q", spec)
		}
		spec = m
	}

	parts := strings.Fields(spec)
	if len(parts) != 5 {
		return nil, fmt.Errorf("expected 5 fields, got %d", len(parts))
	}

	s := &Schedule{Expr: expr}
	sets := [5]*uint64{&s.Minute, &s.Hour, &s.Dom, &s.Month, &s.Dow}
	for i, p := range parts {
		bits, err := parseField(p, fields[i])
		if err != nil {
			return nil, fmt.Errorf("%s: %w", fields[i].name, err)
		}
		*sets[i] = bits
		s.raw[i] = p
	}

	// 7 is an alias for Sunday.
	if s.Dow&(1<<7) != 0 {
		s.Dow = s.Dow&^(1<<7) | 1
	}
	s.domStar = parts[2] == "*" || parts[2] == "?"
	s.dowStar = parts[4] == "*" || parts[4] == "?"
	return s, nil
}

func parseField(s string, f field) (uint64, error) {
	var bits uint64
	for _, item := range strings.Split(s, ",") {
		b, err := parseItem(item, f)
		if err != nil {
			return 0, err
		}
		bits |= b
	}
	return bits, nil
}

func parseItem(item string, f field) (uint64, error) {
	if item == "" {
		return 0, fmt.Errorf("empty list item")
	}

	rangePart, step := item, 1
	if i := strings.IndexByte(item, '/'); i >= 0 {
		rangePart = item[:i]
		n, err := strconv.Atoi(item[i+1:])
		if err != nil || n <= 0 {
			return 0, fmt.Errorf("invalid step in %q", item)
		}
		step = n
	}

	var lo, hi int
	switch {
	case rangePart == "*" || rangePart == "?":
		lo, hi = f.min, f.max
		if f.max == 7 {
			hi = 6
		}
	case strings.Contains(rangePart, "-"):
		a, b, _ := strings.Cut(rangePart, "-")
		var err error
		if lo, err = parseValue(a, f); err != nil {
			return 0, err
		}
		if hi, err = parseValue(b, f); err != nil {
			return 0, err
		}
		if lo > hi {
			return 0, fmt.Errorf("range %q is backwards", rangePart)
		}
	default:
		v, err := parseValue(rangePart, f)
		if err != nil {
			return 0, err
		}
		lo, hi = v, v
		if step > 1 {
			hi = f.max
		}
	}

	var bits uint64
	for v := lo; v <= hi; v += step {
		bits |= 1 << uint(v)
	}
	return bits, nil
}

func parseValue(s string, f field) (int, error) {
	if v, ok := f.names[strings.ToLower(s)]; ok {
		return v, nil
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid value %q", s)
	}
	if v < f.min || v > f.max {
		return 0, fmt.Errorf("value %d out of range %d-%d", v, f.min, f.max)
	}
	return v, nil
}
