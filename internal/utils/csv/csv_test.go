package csv

import (
	"reflect"
	"strconv"
	"testing"
)

type InputCSVRecord struct {
	coinName     string
	contractAddr string
	from         string
	to           string
	network      string
	rate         float64
}

var testFixture = []InputCSVRecord{
	{
		coinName:     "Voyager Token (VGX)",
		contractAddr: "0x3c4b6e6e1ea3d4863700d7f76b36b7f3d3f13e3d",
		from:         "10/11/2021 00:00",
		to:           "17/11/2021 08:35",
		network:      "ethereum",
		rate:         2.8,
	},
}

func testMapper(csvRow []string) (InputCSVRecord, error) {
	r, err := strconv.ParseFloat(csvRow[5], 64)
	if err != nil {
		return InputCSVRecord{}, err
	}
	return InputCSVRecord{
		coinName:     csvRow[0],
		contractAddr: csvRow[1],
		from:         csvRow[3],
		to:           csvRow[4],
		network:      csvRow[2],
		rate:         r,
	}, nil
}

func TestCSVReadAndParse(t *testing.T) {
	cases := []struct {
		in   string
		want []InputCSVRecord
	}{
		{"testdata/test.csv", testFixture},
	}

	for _, c := range cases {
		got := ReadAndParse(c.in, testMapper)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("CSVReadAndParse(%q) == %v, want %v", c.in, got, c.want)
		}
	}
}
