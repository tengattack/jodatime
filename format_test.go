package jodatime_test

import (
	"strconv"
	"testing"
	"time"

	. "github.com/tengattack/jodatime"
)

var (
	local   *time.Location
	timeNow time.Time
)

func init() {
	local, _ = time.LoadLocation("America/Los_Angeles")
	time.Local = local
	timeNow = time.Now()
}

type FormatTest struct {
	name   string
	format string
	result string
}

var formatTests = []FormatTest{
	{"RubyDate", RubyDate, "Wed Feb 04 21:00:57 -0800 2009"},
	{"RFC822", RFC822, "04 Feb 09 21:00 PST"},
	{"RFC822Z", RFC822Z, "04 Feb 09 21:00 -0800"},
	{"RFC850", RFC850, "Wednesday, 04-Feb-09 21:00:57 PST"},
	{"RFC1123", RFC1123, "Wed, 04 Feb 2009 21:00:57 PST"},
	{"RFC1123Z", RFC1123Z, "Wed, 04 Feb 2009 21:00:57 -0800"},
	{"RFC3339", RFC3339, "2009-02-04T21:00:57-08:00"},
	{"RFC3339Nano", RFC3339Nano, "2009-02-04T21:00:57.0123456-08:00"},
	{"Kitchen", Kitchen, "9:00PM"},
	{"AM/PM", "ha", "9PM"},
	{"two-digit year", "YY MM dd", "09 02 04"},
	// Joda time quotes.
	{"escape for text", "'YYZbca'Y", "YYZbca2009"},
	{"single quote", "''YYYY", "'2009"},
}

func TestFormat(t *testing.T) {
	// The numeric time represents Thu Feb  4 21:00:57.012345600 PST 2009
	time := time.Unix(0, 1233810057012345600)
	for _, test := range formatTests {
		result := Format(time, test.format)
		if result != test.result {
			t.Errorf("%s expected %q got %q", test.name, test.result, result)
		}
	}
}

func BenchmarkFormat(b *testing.B) {
	for n := 0; n < b.N; n++ {
		Format(timeNow, RFC3339)
	}
}

func BenchmarkTimeFormat(b *testing.B) {
	t := timeNow
	for n := 0; n < b.N; n++ {
		t.Format(RFC3339)
	}
}

type ParseTest struct {
	name       string
	format     string
	value      string
	hasTZ      bool // contains a time zone
	hasWD      bool // contains a weekday
	yearSign   int  // sign of year, -1 indicates the year is not present in the format
	fracDigits int  // number of digits of fractional second
}

var parseTests = []ParseTest{
	{"RubyDate", RubyDate, "Thu Feb 04 21:00:57 -0800 2010", true, true, 1, 0},
	{"RFC850", RFC850, "Thursday, 04-Feb-10 21:00:57 PST", true, true, 1, 0},
	{"RFC1123", RFC1123, "Thu, 04 Feb 2010 21:00:57 PST", true, true, 1, 0},
	{"RFC1123", RFC1123, "Thu, 04 Feb 2010 22:00:57 PDT", true, true, 1, 0},
	{"RFC1123Z", RFC1123Z, "Thu, 04 Feb 2010 21:00:57 -0800", true, true, 1, 0},
	{"RFC3339", RFC3339, "2010-02-04T21:00:57-08:00", true, false, 1, 0},
	{"custom: \"YYYY-MM-dd HH:mm:ssZ\"", "YYYY-MM-dd HH:mm:ssZ", "2010-02-04 21:00:57-08", true, false, 1, 0},
	// Optional fractional seconds.
	{"RFC850", RFC850, "Thursday, 04-Feb-10 21:00:57.0123 PST", true, true, 1, 4},
	{"RFC1123", RFC1123, "Thu, 04 Feb 2010 21:00:57.01234 PST", true, true, 1, 5},
	{"RFC1123Z", RFC1123Z, "Thu, 04 Feb 2010 21:00:57.01234 -0800", true, true, 1, 5},
	{"RFC3339", RFC3339, "2010-02-04T21:00:57.012345678-08:00", true, false, 1, 9},
	{"custom: \"YYYY-MM-dd HH:mm:ss\"", "YYYY-MM-dd HH:mm:ss", "2010-02-04 21:00:57.0", false, false, 1, 0},
	// Fractional seconds.
	{"millisecond", "EEE MMM dd HH:mm:ss.SSS YYYY", "Thu Feb 04 21:00:57.012 2010", false, true, 1, 3},
	{"microsecond", "EEE MMM dd HH:mm:ss.SSSSSS YYYY", "Thu Feb 04 21:00:57.012345 2010", false, true, 1, 6},
	{"nanosecond", "EEE MMM dd HH:mm:ss.SSSSSSSSS YYYY", "Thu Feb 04 21:00:57.012345678 2010", false, true, 1, 9},
}

func TestParse(t *testing.T) {
	for _, test := range parseTests {
		time, err := Parse(test.format, test.value)
		if err != nil {
			t.Errorf("%s error: %v", test.name, err)
		} else {
			checkTime(time, &test, t)
		}
	}
}

func checkTime(checkTime time.Time, test *ParseTest, t *testing.T) {
	// The time should be Thu Feb  4 21:00:57 PST 2010
	if test.yearSign >= 0 && test.yearSign*checkTime.Year() != 2010 {
		t.Errorf("%s: bad year: %d not %d", test.name, checkTime.Year(), 2010)
	}
	if checkTime.Month() != time.February {
		t.Errorf("%s: bad month: %s not %s", test.name, checkTime.Month(), time.February)
	}
	if checkTime.Day() != 4 {
		t.Errorf("%s: bad day: %d not %d", test.name, checkTime.Day(), 4)
	}
	if checkTime.Hour() != 21 {
		t.Errorf("%s: bad hour: %d not %d", test.name, checkTime.Hour(), 21)
	}
	if checkTime.Minute() != 0 {
		t.Errorf("%s: bad minute: %d not %d", test.name, checkTime.Minute(), 0)
	}
	if checkTime.Second() != 57 {
		t.Errorf("%s: bad second: %d not %d", test.name, checkTime.Second(), 57)
	}
	// Nanoseconds must be checked against the precision of the input.
	nanosec, err := strconv.ParseUint("012345678"[:test.fracDigits]+"000000000"[:9-test.fracDigits], 10, 0)
	if err != nil {
		panic(err)
	}
	if checkTime.Nanosecond() != int(nanosec) {
		t.Errorf("%s: bad nanosecond: %d not %d", test.name, checkTime.Nanosecond(), nanosec)
	}
	name, offset := checkTime.Zone()
	if test.hasTZ && offset != -28800 {
		t.Errorf("%s: bad tz offset: %s %d not %d", test.name, name, offset, -28800)
	}
	if test.hasWD && checkTime.Weekday() != time.Thursday {
		t.Errorf("%s: bad weekday: %s not %s", test.name, checkTime.Weekday(), time.Thursday)
	}
}

func BenchmarkParse(b *testing.B) {
	s := timeNow.Format(time.RFC3339)
	for n := 0; n < b.N; n++ {
		Parse(s, RFC3339)
	}
}

func BenchmarkTimeParse(b *testing.B) {
	s := timeNow.Format(time.RFC3339)
	for n := 0; n < b.N; n++ {
		time.Parse(s, RFC3339)
	}
}
