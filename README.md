# jodatime

[![GoDoc](https://godoc.org/github.com/tengattack/jodatime?status.svg)](https://godoc.org/github.com/tengattack/jodatime)
[![Build Status](https://travis-ci.org/tengattack/jodatime.svg)](https://travis-ci.org/tengattack/jodatime)
[![Coverage Status](https://coveralls.io/repos/github/tengattack/jodatime/badge.svg?branch=master)](https://coveralls.io/github/tengattack/jodatime?branch=master)
[![Go Report Card](http://goreportcard.com/badge/tengattack/jodatime)](http:/goreportcard.com/report/tengattack/jodatime)

A [Go](https://golang.org/)'s `time.Parse` and `time.Format` replacement supports [joda time](http://joda-time.sourceforge.net/apidocs/org/joda/time/format/DateTimeFormat.html) format.

## Introduction

Golang developers refuse to support arbitrary format of fractional seconds:
[#27746](https://github.com/golang/go/issues/27746), [#26002](https://github.com/golang/go/issues/26002), [#6189](https://github.com/golang/go/issues/6189)

So, we can use this package to parse those fractional seconds not in standard format.

## Usage

```go
package main

import (
	"fmt"
	"time"

	"github.com/tengattack/jodatime"
)

func main() {
	date := jodatime.Format(time.Now(), "YYYY.MM.dd")
	fmt.Println(date)

	dateTime, _ := jodatime.Parse("YYYY-MM-dd HH:mm:ss,SSS", "2018-09-19 19:50:26,208")
	fmt.Println(dateTime.String())
}
```

## Format

[http://joda-time.sourceforge.net/apidocs/org/joda/time/format/DateTimeFormat.html](http://joda-time.sourceforge.net/apidocs/org/joda/time/format/DateTimeFormat.html)

## License

MIT
