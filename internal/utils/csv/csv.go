package csv

import (
	"encoding/csv"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gsrai/go-ionic/internal/types"
)

func Download(fileName string, w http.ResponseWriter, headers []string, content []types.OutputCSVRecord) {
	csv.NewWriter(w)

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment;filename="+fileName)
	w.Header().Set("Transfer-Encoding", "chunked")
	writer := csv.NewWriter(w)
	err := writer.Write(headers)
	if err != nil {
		http.Error(w, "Error sending csv: "+err.Error(), http.StatusInternalServerError)
		return
	}
	for _, row := range content {
		ss := row.ToSlice()
		err := writer.Write(ss)
		if err != nil {
			http.Error(w, "Error sending csv: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
	writer.Flush()
}

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
