package livr

import (
	"errors"
	"regexp"
	"time"
)

var (
	datetimeRe = regexp.MustCompile(`^([0-9]{4}-[0-9]{2}-[0-9]{2})T([0-9]{2}):([0-9]{2})(?::([0-9]{2})(?:\.([0-9]{1,9}))?)?(?:Z|([+-])([0-9]{2}):([0-9]{2}))$`)
	timeRe     = regexp.MustCompile(`^([0-9]{2}):([0-9]{2})(?::([0-9]{2})(?:\.([0-9]{1,9}))?)?$`)
)

// Bounds relative to the current date in UTC.
var relativeDays = map[string]int{"yesterday": -1, "today": 0, "tomorrow": 1}

// isoDate - make sure that validated value is valid date in format "2006-01-02" ("YYYY-MM-DD"),
// {"iso_date": {"min": "today", "max": "2026-12-31"}} with bounds. now is the clock of the
// relative bounds.
func isoDate(now func() time.Time) Builder {
	return func(args ...interface{}) Validation {
		min, max := dayBounds("iso_date", args, false)

		return func(value interface{}, builders ...interface{}) (interface{}, interface{}) {
			if isEmpty(value) {
				return value, nil
			}

			s, ok := value.(string)
			if !ok {
				return nil, notStringError(value, "WRONG_DATE")
			}
			t, ok := parseDate(s)
			if !ok {
				return nil, errors.New("WRONG_DATE")
			}
			if err := checkDayBounds(t, min, max, now, "DATE"); err != nil {
				return nil, err
			}

			return value, nil
		}
	}
}

// isoDatetime - make sure that validated value is a date and a time with an offset,
// "2026-10-29T10:00:00+02:00", and convert it to UTC, "2026-10-29T08:00:00.000Z".
// {"iso_datetime": {"min": "2026-01-01T09:00:00Z", "max": "tomorrow"}} with bounds.
func isoDatetime(now func() time.Time) Builder {
	return func(args ...interface{}) Validation {
		min, max := dayBounds("iso_datetime", args, true)

		return func(value interface{}, builders ...interface{}) (interface{}, interface{}) {
			if isEmpty(value) {
				return value, nil
			}

			s, ok := value.(string)
			if !ok {
				return nil, notStringError(value, "WRONG_DATETIME")
			}
			// An instant outside the years 0000-9999 cannot be written in the output form.
			t, ok := parseDatetime(s)
			if !ok || t.Year() < 0 || t.Year() > 9999 {
				return nil, errors.New("WRONG_DATETIME")
			}
			if err := checkDayBounds(t, min, max, now, "DATETIME"); err != nil {
				return nil, err
			}

			return t.Format("2006-01-02T15:04:05.000Z"), nil
		}
	}
}

// isoTime - make sure that validated value is a time of day "HH:MM", "HH:MM:SS" or "HH:MM:SS.sss",
// {"iso_time": {"min": "09:00", "max": "18:00"}} with bounds; min after max is a range over midnight.
func isoTime(args ...interface{}) Validation {
	options := boundOptions("iso_time", args)
	bounds := make(map[string]time.Duration)
	for name, bound := range options {
		s, _ := bound.(string)
		t, ok := parseTime(s)
		if !ok {
			schemaErrorf("Rule iso_time takes a time as %s", name)
		}
		bounds[name] = t
	}
	min, hasMin := bounds["min"]
	max, hasMax := bounds["max"]

	return func(value interface{}, builders ...interface{}) (interface{}, interface{}) {
		if isEmpty(value) {
			return value, nil
		}

		s, ok := value.(string)
		if !ok {
			return nil, notStringError(value, "WRONG_TIME")
		}
		t, ok := parseTime(s)
		if !ok {
			return nil, errors.New("WRONG_TIME")
		}

		if hasMin && hasMax && min > max {
			// Over midnight: a time in the gap gets the error of the nearer bound.
			if t < min && t > max {
				if t-max < min-t {
					return nil, errors.New("TIME_TOO_LATE")
				}
				return nil, errors.New("TIME_TOO_EARLY")
			}
		} else if hasMin && t < min {
			return nil, errors.New("TIME_TOO_EARLY")
		} else if hasMax && t > max {
			return nil, errors.New("TIME_TOO_LATE")
		}

		return value, nil
	}
}

// notStringError - the error of a date rule for a value that is not a string: the code of the
// rule for a number or a boolean, FORMAT_ERROR for an object or a list.
func notStringError(value interface{}, code string) error {
	if _, ok := stringForm(value); ok {
		return errors.New(code)
	}
	return errors.New("FORMAT_ERROR")
}

// boundOptions - the options of a date rule, an object with "min" and "max" as the first argument.
func boundOptions(name string, args []interface{}) Dictionary {
	args, _ = ruleArgs(args)
	// LIVR 2.0 ignored the arguments of iso_date, so {"iso_date": null} stays the plain iso_date.
	if len(args) == 0 || (name == "iso_date" && len(args) == 1 && args[0] == nil) {
		return nil
	}

	options, ok := args[0].(Dictionary)
	if !ok {
		schemaErrorf("Rule %s takes no arguments or an object with min and max", name)
	}
	for key := range options {
		if key != "min" && key != "max" {
			schemaErrorf("Rule %s got an unknown option %q", name, key)
		}
	}
	return options
}

// dayBound - a bound of iso_date and iso_datetime: a fixed instant, or a day relative to the
// current date. A day is its start as min and its end as max.
type dayBound struct {
	isRelative bool
	days       int
	instant    time.Time
	isMax      bool
}

// at - the instant of the bound for the clock now.
func (b *dayBound) at(now func() time.Time) time.Time {
	if !b.isRelative {
		return b.instant
	}
	y, m, d := now().UTC().Date()
	return dayEdge(time.Date(y, m, d+b.days, 0, 0, 0, 0, time.UTC), b.isMax)
}

// dayEdge - the start of the day, or its end (to milliseconds) for max.
func dayEdge(day time.Time, isMax bool) time.Time {
	if isMax {
		return day.Add(24*time.Hour - time.Millisecond)
	}
	return day
}

// dayBounds - the min and max bounds of iso_date and iso_datetime (with datetimes), nil for none.
func dayBounds(name string, args []interface{}, withDatetimes bool) (*dayBound, *dayBound) {
	options := boundOptions(name, args)
	var bounds [2]*dayBound
	for i, option := range []string{"min", "max"} {
		raw, ok := options[option]
		if !ok {
			continue
		}

		s, _ := raw.(string)
		b := &dayBound{isMax: option == "max"}
		if days, ok := relativeDays[s]; ok {
			b.isRelative, b.days = true, days
		} else if day, ok := parseDate(s); ok {
			b.instant = dayEdge(day, b.isMax)
		} else if t, ok := parseDatetime(s); ok && withDatetimes {
			b.instant = t
		} else {
			schemaErrorf(`Rule %s takes a date, "today", "yesterday" or "tomorrow" as %s`, name, option)
		}
		bounds[i] = b
	}

	// No value could pass. A fixed and a relative bound depend on the day, they are not checked.
	min, max := bounds[0], bounds[1]
	if min != nil && max != nil && min.isRelative == max.isRelative &&
		(min.days > max.days || min.instant.After(max.instant)) {
		schemaErrorf("Rule %s got min later than max", name)
	}
	return min, max
}

// checkDayBounds - the bound error of the instant t for the date rule with the code prefix.
func checkDayBounds(t time.Time, min, max *dayBound, now func() time.Time, prefix string) error {
	if min != nil && t.Before(min.at(now)) {
		return errors.New(prefix + "_TOO_EARLY")
	}
	if max != nil && t.After(max.at(now)) {
		return errors.New(prefix + "_TOO_LATE")
	}
	return nil
}

// parseDate - a date YYYY-MM-DD as the start of the day in UTC.
func parseDate(s string) (time.Time, bool) {
	if len(s) != 10 || s[4] != '-' || s[7] != '-' {
		return time.Time{}, false
	}
	year, ok1 := parseDigits(s[0:4])
	month, ok2 := parseDigits(s[5:7])
	day, ok3 := parseDigits(s[8:10])
	if !ok1 || !ok2 || !ok3 || month < 1 || month > 12 || day < 1 {
		return time.Time{}, false
	}

	// time.Date normalizes a day after the end of the month ("02-30" is "03-02").
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return t, t.Day() == day
}

// parseDatetime - a datetime with an offset as an instant, cut to milliseconds.
func parseDatetime(s string) (time.Time, bool) {
	m := datetimeRe.FindStringSubmatch(s)
	if m == nil {
		return time.Time{}, false
	}
	day, ok := parseDate(m[1])
	hours, _ := parseDigits(m[2])
	minutes, _ := parseDigits(m[3])
	seconds, _ := parseDigits(m[4])
	ms, _ := parseDigits((m[5] + "000")[:3])
	if !ok || hours > 23 || minutes > 59 || seconds > 59 {
		return time.Time{}, false
	}
	t := day.Add(time.Duration(((hours*60+minutes)*60+seconds)*1000+ms) * time.Millisecond)

	if m[6] != "" {
		offsetHours, _ := parseDigits(m[7])
		offsetMinutes, _ := parseDigits(m[8])
		if offsetHours > 23 || offsetMinutes > 59 {
			return time.Time{}, false
		}
		offset := time.Duration(offsetHours*60+offsetMinutes) * time.Minute
		if m[6] == "+" {
			offset = -offset
		}
		t = t.Add(offset)
	}
	return t, true
}

// parseTime - a time of day as the time since midnight, all nine fraction digits count.
func parseTime(s string) (time.Duration, bool) {
	m := timeRe.FindStringSubmatch(s)
	if m == nil {
		return 0, false
	}
	hours, _ := parseDigits(m[1])
	minutes, _ := parseDigits(m[2])
	seconds, _ := parseDigits(m[3])
	ns, _ := parseDigits((m[4] + "000000000")[:9])
	if hours > 23 || minutes > 59 || seconds > 59 {
		return 0, false
	}
	return time.Duration((hours*60+minutes)*60+seconds)*time.Second + time.Duration(ns), true
}

// parseDigits - the number of ASCII digits, 0 for "".
func parseDigits(s string) (int, bool) {
	n := 0
	for i := 0; i < len(s); i++ {
		if !isDigit(s[i]) {
			return 0, false
		}
		n = n*10 + int(s[i]-'0')
	}
	return n, true
}
