# cronx

A small command-line tool that explains a cron expression in plain English and
lists when it will run next. No dependencies beyond the Go standard library.

## Build

```
go build -o cronx .
```

## Usage

```
cronx [flags] <expression>
```

Flags must come before the expression.

| flag     | meaning                                              |
|----------|------------------------------------------------------|
| `-n`     | number of upcoming runs to show (default 5)          |
| `-from`  | start time instead of now (`2026-03-01 09:30`, RFC 3339, or a date) |
| `-tz`    | IANA time zone to evaluate in (default local)        |
| `-check` | exit 0 if the schedule fires at the given time, 1 otherwise |
| `-q`     | print only RFC 3339 timestamps, one per line         |

```
$ cronx -tz UTC -from "2026-10-20 10:00" -n 3 '*/15 9-17 * * mon-fri'
Every 15 minutes, during hours 09 through 17, on weekdays

  1. Tue 2026-10-20 10:15 UTC  (in 15m)
  2. Tue 2026-10-20 10:30 UTC  (in 30m)
  3. Tue 2026-10-20 10:45 UTC  (in 45m)

$ cronx -check "2026-10-24 09:30" -tz UTC '30 9 * * 1-5'
no: does not fire at Sat 2026-10-24 09:30 UTC
```

## Supported syntax

Standard five fields: minute, hour, day of month, month, day of week.

- `*`, lists (`1,15`), ranges (`9-17`), steps (`*/10`, `5-55/10`, `5/20`)
- month and weekday names (`jan`, `mon-fri`); `7` is also Sunday
- macros: `@yearly`, `@annually`, `@monthly`, `@weekly`, `@daily`, `@midnight`, `@hourly`
- when both day-of-month and day-of-week are restricted, the job runs if
  either matches, as classic cron does

Times that fall inside a DST spring-forward gap are skipped; a time repeated
by a fall-back shift fires on both occurrences.

## Tests

```
go test ./...
```
