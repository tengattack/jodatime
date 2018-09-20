package jodatime

import (
	"errors"
	"time"
)

// Formats
const (
	RubyDate    = "EEE MMM dd HH:mm:ss Z YYYY"
	RFC822      = "dd MMM YY HH:mm ZZZ"
	RFC822Z     = "dd MMM YY HH:mm Z" // RFC822 with numeric zone
	RFC850      = "EEEE, dd-MMM-YY HH:mm:ss ZZZ"
	RFC1123     = "EEE, dd MMM YYYY HH:mm:ss ZZZ"
	RFC1123Z    = "EEE, dd MMM YYYY HH:mm:ss Z" // RFC1123 with numeric zone
	RFC3339     = "YYYY-MM-ddTHH:mm:ssZZ"
	RFC3339Nano = "YYYY-MM-ddTHH:mm:ss.SSSSSSSSSZZ"
	Kitchen     = "h:mma"
)

const (
	_                        = iota
	stdLongMonth             = iota + stdNeedDate  // "January"
	stdMonth                                       // "Jan"
	stdNumMonth                                    // "1"
	stdZeroMonth                                   // "01"
	stdLongWeekDay                                 // "Monday"
	stdWeekDay                                     // "Mon"
	stdDay                                         // "2"
	stdUnderDay                                    // "_2"
	stdZeroDay                                     // "02"
	stdHour                  = iota + stdNeedClock // "15"
	stdHour12                                      // "3"
	stdZeroHour12                                  // "03"
	stdMinute                                      // "4"
	stdZeroMinute                                  // "04"
	stdSecond                                      // "5"
	stdZeroSecond                                  // "05"
	stdLongYear              = iota + stdNeedDate  // "2006"
	stdYear                                        // "06"
	stdPM                    = iota + stdNeedClock // "PM"
	stdpm                                          // "pm"
	stdTZ                    = iota                // "MST"
	stdISO8601TZ                                   // "Z0700"  // prints Z for UTC
	stdISO8601SecondsTZ                            // "Z070000"
	stdISO8601ShortTZ                              // "Z07"
	stdISO8601ColonTZ                              // "Z07:00" // prints Z for UTC
	stdISO8601ColonSecondsTZ                       // "Z07:00:00"
	stdNumTZ                                       // "-0700"  // always numeric
	stdNumSecondsTz                                // "-070000"
	stdNumShortTZ                                  // "-07"    // always numeric
	stdNumColonTZ                                  // "-07:00" // always numeric
	stdNumColonSecondsTZ                           // "-07:00:00"
	stdFracSecond0                                 // ".0", ".00", ... , trailing zeros included
	stdFracSecond9                                 // ".9", ".99", ..., trailing zeros omitted

	stdNeedDate  = 1 << 8             // need month, day, year
	stdNeedClock = 2 << 8             // need hour, minute, second
	stdArgShift  = 16                 // extra argument in high bits, above low stdArgShift
	stdMask      = 1<<stdArgShift - 1 // mask out argument
)

// std0x records the std values for "01", "02", ..., "06".
var std0x = [...]int{stdZeroMonth, stdZeroDay, stdZeroHour12, stdZeroMinute, stdZeroSecond, stdYear}

// startsWithLowerCase reports whether the string has a lower-case letter at the beginning.
// Its purpose is to prevent matching strings like "Month" when looking for "Mon".
func startsWithLowerCase(str string) bool {
	if len(str) == 0 {
		return false
	}
	c := str[0]
	return 'a' <= c && c <= 'z'
}

// nextStdChunk finds the first occurrence of a std string in
// layout and returns the text before, the std string, and the text after.
func nextStdChunk(layout string) (prefix string, std int, suffix string) {
	layoutLength := len(layout)
	for i := 0; i < layoutLength; i++ {
		switch r := layout[i]; r {
		case 'h':
			j := 1
			for ; i+j < layoutLength; j++ {
				if layout[i+j] != r {
					break
				}
			}
			switch j {
			case 1: // d
				return layout[0:i], stdHour12, layout[i+j:]
			default:
				return layout[0:i], stdZeroHour12, layout[i+j:]
			}
		case 'H':
			j := 1
			for ; i+j < layoutLength; j++ {
				if layout[i+j] != r {
					break
				}
			}
			return layout[0:i], stdHour, layout[i+j:]
		case 'm':
			j := 1
			for ; i+j < layoutLength; j++ {
				if layout[i+j] != r {
					break
				}
			}
			switch j {
			case 1: // d
				return layout[0:i], stdMinute, layout[i+j:]
			default:
				return layout[0:i], stdZeroMinute, layout[i+j:]
			}
		case 's':
			j := 1
			for ; i+j < layoutLength; j++ {
				if layout[i+j] != r {
					break
				}
			}
			switch j {
			case 1: // d
				return layout[0:i], stdSecond, layout[i+j:]
			default:
				return layout[0:i], stdZeroSecond, layout[i+j:]
			}
		case 'd':
			j := 1
			for ; i+j < layoutLength; j++ {
				if layout[i+j] != r {
					break
				}
			}
			switch j {
			case 1: // d
				return layout[0:i], stdDay, layout[i+j:]
			default:
				return layout[0:i], stdZeroDay, layout[i+j:]
			}
		case 'E':
			j := 1
			for ; i+j < layoutLength; j++ {
				if layout[i+j] != r {
					break
				}
			}
			switch j {
			case 1, 2, 3: // d
				return layout[0:i], stdWeekDay, layout[i+j:]
			default:
				return layout[0:i], stdLongWeekDay, layout[i+j:]
			}
		case 'M':
			j := 1
			for ; i+j < layoutLength; j++ {
				if layout[i+j] != r {
					break
				}
			}

			switch j {
			case 1: // d
				return layout[0:i], stdNumMonth, layout[i+j:]
			case 2:
				return layout[0:i], stdZeroMonth, layout[i+j:]
			case 3:
				return layout[0:i], stdMonth, layout[i+j:]
			case 4:
				return layout[0:i], stdLongMonth, layout[i+j:]
			}

			i = i + j - 1
		case 'Y', 'y', 'x':
			j := 1
			for ; i+j < layoutLength; j++ {
				if layout[i+j] != r {
					break
				}
			}
			switch j {
			case 2: // d
				return layout[0:i], stdYear, layout[i+j:]
			default: // dd
				return layout[0:i], stdLongYear, layout[i+j:]
			}
		case 'S':
			j := 1
			for ; i+j < layoutLength; j++ {
				if layout[i+j] != r {
					break
				}
			}

			switch j {
			case 1, 2, 3:
				return layout[0:i], stdFracSecond0 | (j << stdArgShift), layout[i+j:]
			default:
				return layout[0:i], stdFracSecond9 | (j << stdArgShift), layout[i+j:]
			}
		case 'a':
			j := 1
			for ; i+j < layoutLength; j++ {
				if layout[i+j] != r {
					break
				}
			}

			return layout[0:i], stdPM, layout[i+j:]
		case 'Z':
			j := 1
			for ; i+j < layoutLength; j++ {
				if layout[i+j] != r {
					break
				}
			}

			switch j {
			case 1: // d
				return layout[0:i], stdNumTZ, layout[i+j:]
			case 2: // d
				return layout[0:i], stdNumColonTZ, layout[i+j:]
			default: // time zone id
				return layout[0:i], stdTZ, layout[i+j:]
			}
		case '\'': // ' (text delimiter)  or '' (real quote)

			// real quote
			if i+1 < layoutLength && layout[i+1] == r {
				layout = layout[0:i] + layout[i+1:]
				layoutLength = layoutLength - 1
				continue
			}

			tmp := []byte{}
			j := 1
			for ; i+j < layoutLength; j++ {
				if layout[i+j] != r {
					tmp = append(tmp, layout[i+j])
					continue
				}
				break
			}
			layout = layout[0:i] + string(tmp) + layout[i+j+1:]
			i = i + j - 2
			layoutLength = layoutLength - 2
		}
	}
	return layout, 0, ""
}

var longDayNames = []string{
	"Sunday",
	"Monday",
	"Tuesday",
	"Wednesday",
	"Thursday",
	"Friday",
	"Saturday",
}

var shortDayNames = []string{
	"Sun",
	"Mon",
	"Tue",
	"Wed",
	"Thu",
	"Fri",
	"Sat",
}

var shortMonthNames = []string{
	"Jan",
	"Feb",
	"Mar",
	"Apr",
	"May",
	"Jun",
	"Jul",
	"Aug",
	"Sep",
	"Oct",
	"Nov",
	"Dec",
}

var longMonthNames = []string{
	"January",
	"February",
	"March",
	"April",
	"May",
	"June",
	"July",
	"August",
	"September",
	"October",
	"November",
	"December",
}

// match reports whether s1 and s2 match ignoring case.
// It is assumed s1 and s2 are the same length.
func match(s1, s2 string) bool {
	for i := 0; i < len(s1); i++ {
		c1 := s1[i]
		c2 := s2[i]
		if c1 != c2 {
			// Switch to lower-case; 'a'-'A' is known to be a single bit.
			c1 |= 'a' - 'A'
			c2 |= 'a' - 'A'
			if c1 != c2 || c1 < 'a' || c1 > 'z' {
				return false
			}
		}
	}
	return true
}

func lookup(tab []string, val string) (int, string, error) {
	for i, v := range tab {
		if len(val) >= len(v) && match(val[0:len(v)], v) {
			return i, val[len(v):], nil
		}
	}
	return -1, val, errBad
}

// appendInt appends the decimal form of x to b and returns the result.
// If the decimal form (excluding sign) is shorter than width, the result is padded with leading 0's.
// Duplicates functionality in strconv, but avoids dependency.
func appendInt(b []byte, x int, width int) []byte {
	u := uint(x)
	if x < 0 {
		b = append(b, '-')
		u = uint(-x)
	}

	// Assemble decimal in reverse order.
	var buf [20]byte
	i := len(buf)
	for u >= 10 {
		i--
		q := u / 10
		buf[i] = byte('0' + u - q*10)
		u = q
	}
	i--
	buf[i] = byte('0' + u)

	// Add 0-padding.
	for w := len(buf) - i; w < width; w++ {
		b = append(b, '0')
	}

	return append(b, buf[i:]...)
}

// Never printed, just needs to be non-nil for return by atoi.
var atoiError = errors.New("time: invalid number")

// Duplicates functionality in strconv, but avoids dependency.
func atoi(s string) (x int, err error) {
	neg := false
	if s != "" && (s[0] == '-' || s[0] == '+') {
		neg = s[0] == '-'
		s = s[1:]
	}
	q, rem, err := leadingInt(s)
	x = int(q)
	if err != nil || rem != "" {
		return 0, atoiError
	}
	if neg {
		x = -x
	}
	return x, nil
}

// formatNano appends a fractional second, as nanoseconds, to b
// and returns the result.
func formatNano(b []byte, nanosec uint, n int, trim bool) []byte {
	u := nanosec
	var buf [9]byte
	for start := len(buf); start > 0; {
		start--
		buf[start] = byte(u%10 + '0')
		u /= 10
	}

	if n > 9 {
		n = 9
	}
	if trim {
		for n > 0 && buf[n-1] == '0' {
			n--
		}
		if n == 0 {
			return b
		}
	}
	return append(b, buf[:n]...)
}

// Format returns a textual representation of the time value formatted
// according to layout, which defines the format by showing how the reference
// time, defined to be
//  http://joda-time.sourceforge.net/apidocs/org/joda/time/format/DateTimeFormat.html
//
// Predefined layouts RFC3339 and others describe standard
// and convenient representations of the reference time.
func Format(t time.Time, layout string) string {
	const bufSize = 64
	var b []byte
	max := len(layout) + 10
	if max < bufSize {
		var buf [bufSize]byte
		b = buf[:0]
	} else {
		b = make([]byte, 0, max)
	}
	b = AppendFormat(t, b, layout)
	return string(b)
}

// AppendFormat is like Format but appends the textual
// representation to b and returns the extended buffer.
func AppendFormat(t time.Time, b []byte, layout string) []byte {
	var (
		name, offset, abs = locabs(t)

		year  int = -1
		month time.Month
		day   int
		hour  int = -1
		min   int
		sec   int
	)
	// Each iteration generates one std value.
	for layout != "" {
		prefix, std, suffix := nextStdChunk(layout)
		if prefix != "" {
			b = append(b, prefix...)
		}
		if std == 0 {
			break
		}
		layout = suffix

		// Compute year, month, day if needed.
		if year < 0 && std&stdNeedDate != 0 {
			year, month, day, _ = absDate(abs, true)
		}

		// Compute hour, minute, second if needed.
		if hour < 0 && std&stdNeedClock != 0 {
			hour, min, sec = absClock(abs)
		}

		switch std & stdMask {
		case stdYear:
			y := year
			if y < 0 {
				y = -y
			}
			b = appendInt(b, y%100, 2)
		case stdLongYear:
			b = appendInt(b, year, 4)
		case stdMonth:
			b = append(b, month.String()[:3]...)
		case stdLongMonth:
			m := month.String()
			b = append(b, m...)
		case stdNumMonth:
			b = appendInt(b, int(month), 0)
		case stdZeroMonth:
			b = appendInt(b, int(month), 2)
		case stdWeekDay:
			b = append(b, absWeekday(abs).String()[:3]...)
		case stdLongWeekDay:
			s := absWeekday(abs).String()
			b = append(b, s...)
		case stdDay:
			b = appendInt(b, day, 0)
		case stdUnderDay:
			if day < 10 {
				b = append(b, ' ')
			}
			b = appendInt(b, day, 0)
		case stdZeroDay:
			b = appendInt(b, day, 2)
		case stdHour:
			b = appendInt(b, hour, 2)
		case stdHour12:
			// Noon is 12PM, midnight is 12AM.
			hr := hour % 12
			if hr == 0 {
				hr = 12
			}
			b = appendInt(b, hr, 0)
		case stdZeroHour12:
			// Noon is 12PM, midnight is 12AM.
			hr := hour % 12
			if hr == 0 {
				hr = 12
			}
			b = appendInt(b, hr, 2)
		case stdMinute:
			b = appendInt(b, min, 0)
		case stdZeroMinute:
			b = appendInt(b, min, 2)
		case stdSecond:
			b = appendInt(b, sec, 0)
		case stdZeroSecond:
			b = appendInt(b, sec, 2)
		case stdPM:
			if hour >= 12 {
				b = append(b, "PM"...)
			} else {
				b = append(b, "AM"...)
			}
		case stdpm:
			if hour >= 12 {
				b = append(b, "pm"...)
			} else {
				b = append(b, "am"...)
			}
		case stdISO8601TZ, stdISO8601ColonTZ, stdISO8601SecondsTZ, stdISO8601ShortTZ, stdISO8601ColonSecondsTZ, stdNumTZ, stdNumColonTZ, stdNumSecondsTz, stdNumShortTZ, stdNumColonSecondsTZ:
			// Ugly special case. We cheat and take the "Z" variants
			// to mean "the time zone as formatted for ISO 8601".
			if offset == 0 && (std == stdISO8601TZ || std == stdISO8601ColonTZ || std == stdISO8601SecondsTZ || std == stdISO8601ShortTZ || std == stdISO8601ColonSecondsTZ) {
				b = append(b, 'Z')
				break
			}
			zone := offset / 60 // convert to minutes
			absoffset := offset
			if zone < 0 {
				b = append(b, '-')
				zone = -zone
				absoffset = -absoffset
			} else {
				b = append(b, '+')
			}
			b = appendInt(b, zone/60, 2)
			if std == stdISO8601ColonTZ || std == stdNumColonTZ || std == stdISO8601ColonSecondsTZ || std == stdNumColonSecondsTZ {
				b = append(b, ':')
			}
			if std != stdNumShortTZ && std != stdISO8601ShortTZ {
				b = appendInt(b, zone%60, 2)
			}

			// append seconds if appropriate
			if std == stdISO8601SecondsTZ || std == stdNumSecondsTz || std == stdNumColonSecondsTZ || std == stdISO8601ColonSecondsTZ {
				if std == stdNumColonSecondsTZ || std == stdISO8601ColonSecondsTZ {
					b = append(b, ':')
				}
				b = appendInt(b, absoffset%60, 2)
			}

		case stdTZ:
			if name != "" {
				b = append(b, name...)
				break
			}
			// No time zone known for this time, but we must print one.
			// Use the -0700 format.
			zone := offset / 60 // convert to minutes
			if zone < 0 {
				b = append(b, '-')
				zone = -zone
			} else {
				b = append(b, '+')
			}
			b = appendInt(b, zone/60, 2)
			b = appendInt(b, zone%60, 2)
		case stdFracSecond0, stdFracSecond9:
			b = formatNano(b, uint(t.Nanosecond()), std>>stdArgShift, std&stdMask == stdFracSecond9)
		}
	}
	return b
}

// isDigit reports whether s[i] is in range and is a decimal digit.
func isDigit(s string, i int) bool {
	if len(s) <= i {
		return false
	}
	c := s[i]
	return '0' <= c && c <= '9'
}

var errBad = errors.New("bad value for field") // placeholder not passed to user

// getnum parses s[0:1] or s[0:2] (fixed forces the latter)
// as a decimal integer and returns the integer and the
// remainder of the string.
func getnum(s string, fixed bool) (int, string, error) {
	if !isDigit(s, 0) {
		return 0, s, errBad
	}
	if !isDigit(s, 1) {
		if fixed {
			return 0, s, errBad
		}
		return int(s[0] - '0'), s[1:], nil
	}
	return int(s[0]-'0')*10 + int(s[1]-'0'), s[2:], nil
}

func cutspace(s string) string {
	for len(s) > 0 && s[0] == ' ' {
		s = s[1:]
	}
	return s
}

// skip removes the given prefix from value,
// treating runs of space characters as equivalent.
func skip(value, prefix string) (string, error) {
	for len(prefix) > 0 {
		if prefix[0] == ' ' {
			if len(value) > 0 && value[0] != ' ' {
				return value, errBad
			}
			prefix = cutspace(prefix)
			value = cutspace(value)
			continue
		}
		if len(value) == 0 || value[0] != prefix[0] {
			return value, errBad
		}
		prefix = prefix[1:]
		value = value[1:]
	}
	return value, nil
}

// Parse parses time string with joda format:
// http://joda-time.sourceforge.net/apidocs/org/joda/time/format/DateTimeFormat.html
func Parse(layout, value string) (time.Time, error) {
	return parse(layout, value, time.UTC, time.Local)
}

// ParseInLocation is like Parse but differs in two important ways.
// First, in the absence of time zone information, Parse interprets a time as UTC;
// ParseInLocation interprets the time as in the given location.
// Second, when given a zone offset or abbreviation, Parse tries to match it
// against the Local location; ParseInLocation uses the given location.
func ParseInLocation(layout, value string, loc *time.Location) (time.Time, error) {
	return parse(layout, value, loc, loc)
}

func parse(layout, value string, defaultLocation, local *time.Location) (time.Time, error) {
	alayout, avalue := layout, value
	rangeErrString := "" // set if a value is out of range
	amSet := false       // do we need to subtract 12 from the hour for midnight?
	pmSet := false       // do we need to add 12 to the hour?

	// Time being constructed.
	var (
		year       int
		month      int = 1 // January
		day        int = 1
		hour       int
		min        int
		sec        int
		nsec       int
		z          *time.Location
		zoneOffset int = -1
		zoneName   string
	)

	// Each iteration processes one std value.
	for {
		var err error
		prefix, std, suffix := nextStdChunk(layout)
		stdstr := layout[len(prefix) : len(layout)-len(suffix)]
		value, err = skip(value, prefix)
		if err != nil {
			return time.Time{}, &time.ParseError{alayout, avalue, prefix, value, ""}
		}
		if std == 0 {
			if len(value) != 0 {
				return time.Time{}, &time.ParseError{alayout, avalue, "", value, ": extra text: " + value}
			}
			break
		}
		layout = suffix
		var p string
		switch std & stdMask {
		case stdYear:
			if len(value) < 2 {
				err = errBad
				break
			}
			p, value = value[0:2], value[2:]
			year, err = atoi(p)
			if year >= 69 { // Unix time starts Dec 31 1969 in some time zones
				year += 1900
			} else {
				year += 2000
			}
		case stdLongYear:
			if len(value) < 4 || !isDigit(value, 0) {
				err = errBad
				break
			}
			p, value = value[0:4], value[4:]
			year, err = atoi(p)
		case stdMonth:
			month, value, err = lookup(shortMonthNames, value)
			month++
		case stdLongMonth:
			month, value, err = lookup(longMonthNames, value)
			month++
		case stdNumMonth, stdZeroMonth:
			month, value, err = getnum(value, std == stdZeroMonth)
			if month <= 0 || 12 < month {
				rangeErrString = "month"
			}
		case stdWeekDay:
			// Ignore weekday except for error checking.
			_, value, err = lookup(shortDayNames, value)
		case stdLongWeekDay:
			_, value, err = lookup(longDayNames, value)
		case stdDay, stdUnderDay, stdZeroDay:
			if std == stdUnderDay && len(value) > 0 && value[0] == ' ' {
				value = value[1:]
			}
			day, value, err = getnum(value, std == stdZeroDay)
			if day < 0 {
				// Note that we allow any one- or two-digit day here.
				rangeErrString = "day"
			}
		case stdHour:
			hour, value, err = getnum(value, false)
			if hour < 0 || 24 <= hour {
				rangeErrString = "hour"
			}
		case stdHour12, stdZeroHour12:
			hour, value, err = getnum(value, std == stdZeroHour12)
			if hour < 0 || 12 < hour {
				rangeErrString = "hour"
			}
		case stdMinute, stdZeroMinute:
			min, value, err = getnum(value, std == stdZeroMinute)
			if min < 0 || 60 <= min {
				rangeErrString = "minute"
			}
		case stdSecond, stdZeroSecond:
			sec, value, err = getnum(value, std == stdZeroSecond)
			if sec < 0 || 60 <= sec {
				rangeErrString = "second"
			}
			// Special case: do we have a fractional second but no
			// fractional second in the format?
			if len(value) >= 2 && value[0] == '.' && isDigit(value, 1) {
				_, std, _ = nextStdChunk(layout)
				std &= stdMask
				if std == stdFracSecond0 || std == stdFracSecond9 {
					// Fractional second in the layout; proceed normally
					break
				}
				// No fractional second in the layout but we have one in the input.
				n := 2
				for ; n < len(value) && isDigit(value, n); n++ {
				}
				nsec, rangeErrString, err = parseNanoseconds(value[1:], n-1) // remove first dot
				value = value[n:]
			}
		case stdPM:
			if len(value) < 2 {
				err = errBad
				break
			}
			p, value = value[0:2], value[2:]
			switch p {
			case "PM":
				pmSet = true
			case "AM":
				amSet = true
			default:
				err = errBad
			}
		case stdpm:
			if len(value) < 2 {
				err = errBad
				break
			}
			p, value = value[0:2], value[2:]
			switch p {
			case "pm":
				pmSet = true
			case "am":
				amSet = true
			default:
				err = errBad
			}
		case stdISO8601TZ, stdISO8601ColonTZ, stdISO8601SecondsTZ, stdISO8601ShortTZ, stdISO8601ColonSecondsTZ, stdNumTZ, stdNumShortTZ, stdNumColonTZ, stdNumSecondsTz, stdNumColonSecondsTZ:
			if (std == stdISO8601TZ || std == stdISO8601ShortTZ || std == stdISO8601ColonTZ) && len(value) >= 1 && value[0] == 'Z' {
				value = value[1:]
				z = time.UTC
				break
			}
			if std == stdNumTZ {
				if len(value) == 3 {
					// convert to short tz format
					std = stdNumShortTZ
				}
			}
			var sign, hour, min, seconds string
			if std == stdISO8601ColonTZ || std == stdNumColonTZ {
				if len(value) < 6 {
					err = errBad
					break
				}
				if value[3] != ':' {
					err = errBad
					break
				}
				sign, hour, min, seconds, value = value[0:1], value[1:3], value[4:6], "00", value[6:]
			} else if std == stdNumShortTZ || std == stdISO8601ShortTZ {
				if len(value) < 3 {
					err = errBad
					break
				}
				sign, hour, min, seconds, value = value[0:1], value[1:3], "00", "00", value[3:]
			} else if std == stdISO8601ColonSecondsTZ || std == stdNumColonSecondsTZ {
				if len(value) < 9 {
					err = errBad
					break
				}
				if value[3] != ':' || value[6] != ':' {
					err = errBad
					break
				}
				sign, hour, min, seconds, value = value[0:1], value[1:3], value[4:6], value[7:9], value[9:]
			} else if std == stdISO8601SecondsTZ || std == stdNumSecondsTz {
				if len(value) < 7 {
					err = errBad
					break
				}
				sign, hour, min, seconds, value = value[0:1], value[1:3], value[3:5], value[5:7], value[7:]
			} else {
				if len(value) < 5 {
					err = errBad
					break
				}
				sign, hour, min, seconds, value = value[0:1], value[1:3], value[3:5], "00", value[5:]
			}
			var hr, mm, ss int
			hr, err = atoi(hour)
			if err == nil {
				mm, err = atoi(min)
			}
			if err == nil {
				ss, err = atoi(seconds)
			}
			zoneOffset = (hr*60+mm)*60 + ss // offset is in seconds
			switch sign[0] {
			case '+':
			case '-':
				zoneOffset = -zoneOffset
			default:
				err = errBad
			}
		case stdTZ:
			// Does it look like a time zone?
			if len(value) >= 3 && value[0:3] == "UTC" {
				z = time.UTC
				value = value[3:]
				break
			}
			n, ok := parseTimeZone(value)
			if !ok {
				err = errBad
				break
			}
			zoneName, value = value[:n], value[n:]

		case stdFracSecond0:
			// stdFracSecond0 requires the exact number of digits as specified in
			// the layout.
			ndigit := std >> stdArgShift
			if len(value) < ndigit {
				err = errBad
				break
			}
			nsec, rangeErrString, err = parseNanoseconds(value, ndigit)
			value = value[ndigit:]

		case stdFracSecond9:
			if len(value) < 1 || value[0] < '0' || '9' < value[0] {
				// Fractional second omitted.
				break
			}
			// Take any number of digits, even more than asked for,
			// because it is what the stdSecond case would do.
			i := 0
			for i < 9 && i < len(value) && '0' <= value[i] && value[i] <= '9' {
				i++
			}
			nsec, rangeErrString, err = parseNanoseconds(value, i)
			value = value[i:]
		}
		if rangeErrString != "" {
			return time.Time{}, &time.ParseError{alayout, avalue, stdstr, value, ": " + rangeErrString + " out of range"}
		}
		if err != nil {
			return time.Time{}, &time.ParseError{alayout, avalue, stdstr, value, ""}
		}
	}
	if pmSet && hour < 12 {
		hour += 12
	} else if amSet && hour == 12 {
		hour = 0
	}

	// Validate the day of the month.
	if day < 1 || day > daysIn(time.Month(month), year) {
		return time.Time{}, &time.ParseError{alayout, avalue, "", value, ": day out of range"}
	}

	if z != nil {
		return time.Date(year, time.Month(month), day, hour, min, sec, nsec, z), nil
	}

	if zoneOffset != -1 {
		// Create fake zone to record offset.
		z = time.FixedZone(zoneName, zoneOffset)
		// TODO: check UTC and local zone
		t := time.Date(year, time.Month(month), day, hour, min, sec, nsec, z)
		return t, nil
	}

	if zoneName != "" {
		// Look for local zone with the given offset.
		// If that zone was in effect at the given time, use it.
		var err error
		var offset int
		zone, ok := shortIDs[zoneName]
		if ok {
			if zone.name != "" {
				zoneName = zone.name
			} else {
				t := time.Date(year, time.Month(month), day, hour, min, sec, nsec, local)
				_, offset = t.Zone()
				if offset != zone.offset {
					return t.Add(time.Duration(offset-zone.offset) * time.Second), nil
				}
			}
		}
		z, err = time.LoadLocation(zoneName)
		if err == nil {
			t := time.Date(year, time.Month(month), day, hour, min, sec, nsec, z)
			return t, nil
		}

		// Otherwise, create fake zone with unknown offset.
		if len(zoneName) > 3 && zoneName[:3] == "GMT" {
			offset, _ = atoi(zoneName[3:]) // Guaranteed OK by parseGMT.
			offset *= 3600
		}
		z = time.FixedZone(zoneName, offset)
		t := time.Date(year, time.Month(month), day, hour, min, sec, nsec, z)
		return t, nil
	}

	// Otherwise, fall back to default.
	return time.Date(year, time.Month(month), day, hour, min, sec, nsec, defaultLocation), nil
}

// parseTimeZone parses a time zone string and returns its length. Time zones
// are human-generated and unpredictable. We can't do precise error checking.
// On the other hand, for a correct parse there must be a time zone at the
// beginning of the string, so it's almost always true that there's one
// there. We look at the beginning of the string for a run of upper-case letters.
// If there are more than 5, it's an error.
// If there are 4 or 5 and the last is a T, it's a time zone.
// If there are 3, it's a time zone.
// Otherwise, other than special cases, it's not a time zone.
// GMT is special because it can have an hour offset.
func parseTimeZone(value string) (length int, ok bool) {
	if len(value) < 3 {
		return 0, false
	}
	// Special case 1: ChST and MeST are the only zones with a lower-case letter.
	if len(value) >= 4 && (value[:4] == "ChST" || value[:4] == "MeST") {
		return 4, true
	}
	// Special case 2: GMT may have an hour offset; treat it specially.
	if value[:3] == "GMT" {
		length = parseGMT(value)
		return length, true
	}
	// Special Case 3: Some time zones are not named, but have +/-00 format
	if value[0] == '+' || value[0] == '-' {
		length = parseSignedOffset(value)
		return length, true
	}
	// How many upper-case letters are there? Need at least three, at most five.
	var nUpper int
	for nUpper = 0; nUpper < 6; nUpper++ {
		if nUpper >= len(value) {
			break
		}
		if c := value[nUpper]; c < 'A' || 'Z' < c {
			break
		}
	}
	switch nUpper {
	case 0, 1, 2, 6:
		return 0, false
	case 5: // Must end in T to match.
		if value[4] == 'T' {
			return 5, true
		}
	case 4:
		// Must end in T, except one special case.
		if value[3] == 'T' || value[:4] == "WITA" {
			return 4, true
		}
	case 3:
		return 3, true
	}
	return 0, false
}

// parseGMT parses a GMT time zone. The input string is known to start "GMT".
// The function checks whether that is followed by a sign and a number in the
// range -14 through 12 excluding zero.
func parseGMT(value string) int {
	value = value[3:]
	if len(value) == 0 {
		return 3
	}

	return 3 + parseSignedOffset(value)
}

// parseSignedOffset parses a signed timezone offset (e.g. "+03" or "-04").
// The function checks for a signed number in the range -14 through +12 excluding zero.
// Returns length of the found offset string or 0 otherwise
func parseSignedOffset(value string) int {
	sign := value[0]
	if sign != '-' && sign != '+' {
		return 0
	}
	x, rem, err := leadingInt(value[1:])
	if err != nil {
		return 0
	}
	if sign == '-' {
		x = -x
	}
	if x == 0 || x < -14 || 12 < x {
		return 0
	}
	return len(value) - len(rem)
}

func parseNanoseconds(value string, nbytes int) (ns int, rangeErrString string, err error) {
	if ns, err = atoi(value[:nbytes]); err != nil {
		return
	}
	if ns < 0 || 1e9 <= ns {
		rangeErrString = "fractional second"
		return
	}
	// We need nanoseconds, which means scaling by the number
	// of missing digits in the format, maximum length 9. If it's
	// longer than 9, we won't scale.
	scaleDigits := 9 - nbytes
	for i := 0; i < scaleDigits; i++ {
		ns *= 10
	}
	return
}

var errLeadingInt = errors.New("time: bad [0-9]*") // never printed

// leadingInt consumes the leading [0-9]* from s.
func leadingInt(s string) (x int64, rem string, err error) {
	i := 0
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		if x > (1<<63-1)/10 {
			// overflow
			return 0, "", errLeadingInt
		}
		x = x*10 + int64(c) - '0'
		if x < 0 {
			// overflow
			return 0, "", errLeadingInt
		}
	}
	return x, s[i:], nil
}

// leadingFraction consumes the leading [0-9]* from s.
// It is used only for fractions, so does not return an error on overflow,
// it just stops accumulating precision.
func leadingFraction(s string) (x int64, scale float64, rem string) {
	i := 0
	scale = 1
	overflow := false
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		if overflow {
			continue
		}
		if x > (1<<63-1)/10 {
			// It's possible for overflow to give a positive number, so take care.
			overflow = true
			continue
		}
		y := x*10 + int64(c) - '0'
		if y < 0 {
			overflow = true
			continue
		}
		x = y
		scale *= 10
	}
	return x, scale, s[i:]
}
