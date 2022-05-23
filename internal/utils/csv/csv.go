package csv

import (
	"encoding/csv"
	"io"
	"log"
	"os"
)

func ReadAndParse[T any](path string, mapper func([]string) (T, error)) []T {
	var parsedCSVData []T

	header := []string{}

	f, err := os.Open(path)

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	reader := csv.NewReader(f)

	for {
		line, err := reader.Read()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatal(err)
		}

		if len(header) == 0 {
			header = line
			continue
		}

		parsedRow, err := mapper(line)

		if err != nil {
			log.Fatal(err)
		}

		parsedCSVData = append(parsedCSVData, parsedRow)
	}

	return parsedCSVData
}
