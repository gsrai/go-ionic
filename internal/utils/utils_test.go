package utils

import (
	"fmt"
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

func TestToISOString(t *testing.T) {
	cases := []struct {
		in   time.Time
		want string
	}{
		{time.Date(2021, 11, 25, 19, 50, 0, 0, time.UTC), "2021-11-25T19:50:00Z"},
	}

	for _, c := range cases {
		got := ToISOString(c.in)
		if got != c.want {
			t.Errorf("ToISOString(%v) == %v, want %q", c.in, got, c.want)
		}
	}
}

func TestGenFileName(t *testing.T) {
	in := time.Now()
	want := fmt.Sprintf("wallets_%s.csv", in.Format("2006-01-02_15:04:05"))

	got := GenFileName(in)
	if got != want {
		t.Errorf("GenFileName(%v) == %q, want %q", in, got, want)
	}
}
