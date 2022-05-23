package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gsrai/go-ionic/internal/types"
	"github.com/gsrai/go-ionic/internal/utils"
	"github.com/gsrai/go-ionic/internal/utils/csv"
)

const HOST = "127.0.0.1"
const PORT = "8080"
const INPUT_FILE_PATH = "tmp/input.csv"

func main() {
	addr := HOST + ":" + PORT

	http.HandleFunc("/input/load", loadData)

	log.Print("Running server on " + addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func loadData(w http.ResponseWriter, req *http.Request) {
	data := csv.ReadAndParse(INPUT_FILE_PATH, func(csvRow []string) (types.InputCSVRecord, error) {
		rate, err := strconv.ParseFloat(csvRow[5], 64)
		if err != nil {
			return types.InputCSVRecord{}, err
		}

		from, err := utils.ParseDateTime(csvRow[3])
		if err != nil {
			return types.InputCSVRecord{}, err
		}

		to, err := utils.ParseDateTime(csvRow[4])
		if err != nil {
			return types.InputCSVRecord{}, err
		}

		return types.InputCSVRecord{
			CoinName:     csvRow[0],
			ContractAddr: csvRow[1],
			From:         from,
			To:           to,
			Network:      csvRow[2],
			Rate:         rate,
		}, nil
	})

	for idx, datum := range data {
		fmt.Fprintf(w, "row %d: %v\n", idx+1, datum)
	}
}
