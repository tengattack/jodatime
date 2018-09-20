package jodatime

import "time"

// absWeekday is like Weekday but operates on an absolute time.
func absWeekday(abs uint64) time.Weekday {
	// January 1 of the absolute year, like January 1 of 2001, was a Monday.
	sec := (abs + uint64(time.Monday)*secondsPerDay) % secondsPerWeek
	return time.Weekday(int(sec) / secondsPerDay)
}

const (
	secondsPerMinute = 60
	secondsPerHour   = 60 * 60
	secondsPerDay    = 24 * secondsPerHour
	secondsPerWeek   = 7 * secondsPerDay
	daysPer400Years  = 365*400 + 97
	daysPer100Years  = 365*100 + 24
	daysPer4Years    = 365*4 + 1
)

// absClock is like clock but operates on an absolute time.
func absClock(abs uint64) (hour, min, sec int) {
	sec = int(abs % secondsPerDay)
	hour = sec / secondsPerHour
	sec -= hour * secondsPerHour
	min = sec / secondsPerMinute
	sec -= min * secondsPerMinute
	return
}

const (
	// The unsigned zero year for internal calculations.
	// Must be 1 mod 400, and times before it will not compute correctly,
	// but otherwise can be changed at will.
	absoluteZeroYear = -292277022399

	// The year of the zero Time.
	// Assumed by the unixToInternal computation below.
	internalYear = 1

	// Offsets to convert between internal and absolute or Unix times.
	absoluteToInternal int64 = (absoluteZeroYear - internalYear) * 365.2425 * secondsPerDay
	internalToAbsolute       = -absoluteToInternal

	unixToInternal int64 = (1969*365 + 1969/4 - 1969/100 + 1969/400) * secondsPerDay
	internalToUnix int64 = -unixToInternal
)

// absDate is like date but operates on an absolute time.
func absDate(abs uint64, full bool) (year int, month time.Month, day int, yday int) {
	// Split into time and day.
	d := abs / secondsPerDay

	// Account for 400 year cycles.
	n := d / daysPer400Years
	y := 400 * n
	d -= daysPer400Years * n

	// Cut off 100-year cycles.
	// The last cycle has one extra leap year, so on the last day
	// of that year, day / daysPer100Years will be 4 instead of 3.
	// Cut it back down to 3 by subtracting n>>2.
	n = d / daysPer100Years
	n -= n >> 2
	y += 100 * n
	d -= daysPer100Years * n

	// Cut off 4-year cycles.
	// The last cycle has a missing leap year, which does not
	// affect the computation.
	n = d / daysPer4Years
	y += 4 * n
	d -= daysPer4Years * n

	// Cut off years within a 4-year cycle.
	// The last year is a leap year, so on the last day of that year,
	// day / 365 will be 4 instead of 3. Cut it back down to 3
	// by subtracting n>>2.
	n = d / 365
	n -= n >> 2
	y += n
	d -= 365 * n

	year = int(int64(y) + absoluteZeroYear)
	yday = int(d)

	if !full {
		return
	}

	day = yday
	if isLeap(year) {
		// Leap year
		switch {
		case day > 31+29-1:
			// After leap day; pretend it wasn't there.
			day--
		case day == 31+29-1:
			// Leap day.
			month = time.February
			day = 29
			return
		}
	}

	// Estimate month on assumption that every month has 31 days.
	// The estimate may be too low by at most one month, so adjust.
	month = time.Month(day / 31)
	end := int(daysBefore[month+1])
	var begin int
	if day >= end {
		month++
		begin = end
	} else {
		begin = int(daysBefore[month])
	}

	month++ // because January is 1
	day = day - begin + 1
	return
}

// daysBefore[m] counts the number of days in a non-leap year
// before month m begins. There is an entry for m=12, counting
// the number of days before January of next year (365).
var daysBefore = [...]int32{
	0,
	31,
	31 + 28,
	31 + 28 + 31,
	31 + 28 + 31 + 30,
	31 + 28 + 31 + 30 + 31,
	31 + 28 + 31 + 30 + 31 + 30,
	31 + 28 + 31 + 30 + 31 + 30 + 31,
	31 + 28 + 31 + 30 + 31 + 30 + 31 + 31,
	31 + 28 + 31 + 30 + 31 + 30 + 31 + 31 + 30,
	31 + 28 + 31 + 30 + 31 + 30 + 31 + 31 + 30 + 31,
	31 + 28 + 31 + 30 + 31 + 30 + 31 + 31 + 30 + 31 + 30,
	31 + 28 + 31 + 30 + 31 + 30 + 31 + 31 + 30 + 31 + 30 + 31,
}

func daysIn(m time.Month, year int) int {
	if m == time.February && isLeap(year) {
		return 29
	}
	return int(daysBefore[m] - daysBefore[m-1])
}

type zone struct {
	name   string
	offset int
}

// shortIDs is the zone id map
// https://docs.oracle.com/javase/8/docs/api/java/time/ZoneId.html
var shortIDs = map[string]zone{
	"EST": zone{"", -18000},
	"HST": zone{"", -36000},
	"MST": zone{"", -25200},
	"ACT": zone{"Australia/Darwin", 0},
	"AET": zone{"Australia/Sydney", 0},
	"AGT": zone{"America/Argentina/Buenos_Aires", 0},
	"ART": zone{"Africa/Cairo", 0},
	"AST": zone{"America/Anchorage", 0},
	"BET": zone{"America/Sao_Paulo", 0},
	"BST": zone{"Asia/Dhaka", 0},
	"CAT": zone{"Africa/Harare", 0},
	"CNT": zone{"America/St_Johns", 0},
	"CST": zone{"America/Chicago", 0},
	"CTT": zone{"Asia/Shanghai", 0},
	"EAT": zone{"Africa/Addis_Ababa", 0},
	"ECT": zone{"Europe/Paris", 0},
	"IET": zone{"America/Indiana/Indianapolis", 0},
	"IST": zone{"Asia/Kolkata", 0},
	"JST": zone{"Asia/Tokyo", 0},
	"MIT": zone{"Pacific/Apia", 0},
	"NET": zone{"Asia/Yerevan", 0},
	"NST": zone{"Pacific/Auckland", 0},
	"PLT": zone{"Asia/Karachi", 0},
	"PNT": zone{"America/Phoenix", 0},
	"PRT": zone{"America/Puerto_Rico", 0},
	"PST": zone{"America/Los_Angeles", 0},
	"SST": zone{"Pacific/Guadalcanal", 0},
	"VST": zone{"Asia/Ho_Chi_Minh", 0},
	// extra zones
	"PDT": zone{"", -25200},
}

func isLeap(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

func locabs(t time.Time) (name string, offset int, abs uint64) {
	l := t.Location()
	name, offset = t.Zone()
	sec := t.Unix()
	if l != time.UTC {
		sec += int64(offset)
	}
	abs = uint64(sec + (unixToInternal + internalToAbsolute))
	return
}
