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
