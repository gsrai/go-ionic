package utils

import "time"

const INPUT_CSV_TIME_FORMAT = "2/1/2006 15:04"

func ParseDateTime(s string) (time.Time, error) {
	return time.Parse(INPUT_CSV_TIME_FORMAT, s)
}
