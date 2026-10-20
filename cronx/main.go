package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/HisokaDrip/side-projects/cronx/schedule"
)

var timeLayouts = []string{
	time.RFC3339,
	"2006-01-02 15:04",
	"2006-01-02T15:04",
	"2006-01-02",
}

func main() {
	count := flag.Int("n", 5, "number of upcoming runs to show")
	from := flag.String("from", "", "start time (default now), e.g. \"2026-03-01 09:30\"")
	tz := flag.String("tz", "", "IANA time zone, e.g. Europe/Berlin (default local)")
	check := flag.String("check", "", "report whether the schedule fires at this time")
	quiet := flag.Bool("q", false, "print only the run times")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: cronx [flags] <expression>\n\n")
		fmt.Fprintf(os.Stderr, "  cronx '*/15 9-17 * * mon-fri'\n  cronx -n 3 -tz UTC @weekly\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if err := run(strings.Join(flag.Args(), " "), *count, *from, *tz, *check, *quiet); err != nil {
		fmt.Fprintln(os.Stderr, "cronx:", err)
		os.Exit(1)
	}
}

func run(expr string, count int, from, tz, check string, quiet bool) error {
	if expr == "" {
		flag.Usage()
		os.Exit(2)
	}

	loc := time.Local
	if tz != "" {
		var err error
		if loc, err = time.LoadLocation(tz); err != nil {
			return err
		}
	}

	s, err := schedule.Parse(expr)
	if err != nil {
		return err
	}

	if check != "" {
		t, err := parseTime(check, loc)
		if err != nil {
			return err
		}
		if s.Matches(t) {
			fmt.Printf("yes: fires at %s\n", t.Format("Mon 2006-01-02 15:04 MST"))
			return nil
		}
		fmt.Printf("no: does not fire at %s\n", t.Format("Mon 2006-01-02 15:04 MST"))
		os.Exit(1)
	}

	start := time.Now().In(loc)
	if from != "" {
		if start, err = parseTime(from, loc); err != nil {
			return err
		}
	}

	runs := s.Upcoming(start, count)
	if !quiet {
		fmt.Println(s.Describe())
		fmt.Println()
	}
	if len(runs) == 0 {
		return fmt.Errorf("no runs within the next few years")
	}
	for i, t := range runs {
		if quiet {
			fmt.Println(t.Format(time.RFC3339))
			continue
		}
		fmt.Printf("  %d. %s  (%s)\n", i+1, t.Format("Mon 2006-01-02 15:04 MST"), relative(t.Sub(start)))
	}
	return nil
}

func parseTime(s string, loc *time.Location) (time.Time, error) {
	for _, layout := range timeLayouts {
		if t, err := time.ParseInLocation(layout, s, loc); err == nil {
			return t.In(loc), nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse time %q (try \"2006-01-02 15:04\")", s)
}

func relative(d time.Duration) string {
	d = d.Round(time.Minute)
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	mins := int(d.Minutes()) % 60

	var parts []string
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if mins > 0 || len(parts) == 0 {
		parts = append(parts, fmt.Sprintf("%dm", mins))
	}
	return "in " + strings.Join(parts, " ")
}
