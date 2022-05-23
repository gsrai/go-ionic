package utils

import (
	"testing"
	"time"
)

func TestParseDateTime(t *testing.T) {
	cases := []struct {
		in   string
		want time.Time
	}{
		{"25/11/2021 19:50", time.Date(2021, 11, 25, 19, 50, 0, 0, time.UTC)},
	}

	for _, c := range cases {
		got, err := ParseDateTime(c.in)
		if err != nil {
			t.Fatal(err)
		}

		if got != c.want {
			t.Errorf("ParseDateTime(%q) == %v, want %v", c.in, got, c.want)
		}
	}
}
