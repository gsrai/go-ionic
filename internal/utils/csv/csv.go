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
		log.Panic(err)
	}

	defer f.Close()

	reader := csv.NewReader(f)

	for {
		line, err := reader.Read()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Panic(err)
		}

		if len(header) == 0 {
			header = line
			continue
		}

		parsedRow, err := mapper(line)

		if err != nil {
			log.Panic(err)
		}

		parsedCSVData = append(parsedCSVData, parsedRow)
	}

	return parsedCSVData
}

func ReadAndParseP[T any](path string, parseFn func([]string) (T, error), out chan<- T) {
	fd, osErr := os.Open(path)
	if osErr != nil {
		log.Panic(osErr)
	}

	defer fd.Close()
	defer close(out)

	reader := csv.NewReader(fd)
	hasSkippedHeader := false
	for {
		switch line, err := reader.Read(); {
		case err == io.EOF:
			return
		case err != nil:
			log.Panic(err)
		case !hasSkippedHeader:
			hasSkippedHeader = true
			continue
		default:
			if parsedRow, parserErr := parseFn(line); parserErr != nil {
				log.Panic(parserErr)
			} else {
				out <- parsedRow
			}
		}
	}
}
