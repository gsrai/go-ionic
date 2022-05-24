package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gsrai/go-ionic/internal/clients/covalent"
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
	http.HandleFunc("/block/heights", getRecentBlockHeight)
	http.HandleFunc("/eventlog", readEventLog)

	log.Print("Running server on " + addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

type TransferEvent struct {
	From     types.Address
	To       types.Address
	Amount   float64
	CoinName string
}

func readEventLog(w http.ResponseWriter, req *http.Request) {
	res := covalent.GetLogEvents("0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48", 14256555, 14256605, types.ETH)

	var transferEvents []TransferEvent
	for _, item := range res.Data.Items {
		if item.DecodedEvent.Name == "Transfer" {

			from, _ := item.DecodedEvent.Params[0].Value.(string)
			to, _ := item.DecodedEvent.Params[1].Value.(string)
			foo, _ := item.DecodedEvent.Params[2].Value.(string)
			bar, _ := strconv.ParseFloat(foo, 64)

			transferEvents = append(transferEvents, TransferEvent{
				From:     types.Address(from),
				To:       types.Address(to),
				Amount:   bar / math.Pow10(item.ContractDecimals),
				CoinName: item.ContractTickerSymbol,
			})
		}
	}

	fmt.Fprintf(w, "|%-12s|%-6s| %-42s | %-42s |\n", "Amount", "Token", "From", "To")
	for _, t := range transferEvents {
		fmt.Fprintf(w, "|%12.2f|%6s| %s | %s |\n", t.Amount, t.CoinName, t.From, t.To)
	}
}

func getRecentBlockHeight(w http.ResponseWriter, req *http.Request) {
	start := time.Date(2021, 11, 10, 0, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	res := covalent.GetBlockHeights(types.ETH, start, end)
	fmt.Fprintf(w, "block heights: %v\n", res)
}

func loadData(w http.ResponseWriter, req *http.Request) {
	mapperFunc := func(csvRow []string) (types.InputCSVRecord, error) {
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
	}

	data := csv.ReadAndParse(INPUT_FILE_PATH, mapperFunc)

	for idx, datum := range data {
		fmt.Fprintf(w, "row %d: %v\n", idx+1, datum)
	}
}
