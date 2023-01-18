package utils

import (
	"fmt"
	"time"
)

const INPUT_CSV_TIME_FORMAT = "2/1/2006 15:04"

func ParseDateTime(s string) (time.Time, error) {
	return time.Parse(INPUT_CSV_TIME_FORMAT, s)
}

func ToISOString(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

func GenFileName(t time.Time) string {
	return fmt.Sprintf("wallets_%s.csv", t.Format("2006-01-02_15:04:05"))
}

func GetMapKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
