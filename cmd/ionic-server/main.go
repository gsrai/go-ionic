package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gsrai/go-ionic/config"
	"github.com/gsrai/go-ionic/internal/clients/covalent"
	"github.com/gsrai/go-ionic/internal/core"
	t "github.com/gsrai/go-ionic/internal/types"
	"github.com/gsrai/go-ionic/internal/utils"
	"github.com/gsrai/go-ionic/internal/utils/csv"
)

func main() {
	config.Init()
	host, port := config.Get().ServerHost, config.Get().ServerPort
	addr := host + ":" + port

	http.HandleFunc("/", getWallets)

	log.Print("Running server on " + addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func getWallets(w http.ResponseWriter, req *http.Request) {
	defer func(t time.Time) {
		elapsed := time.Since(t)
		fmt.Printf("finished getting wallets in %s 👍\n", elapsed)
	}(time.Now())

	var histories [][]t.TransferEvent
	var uniqueHistories [][]t.CoinTradeInfo

	data := csv.ReadAndParse(config.Get().InputFilePath, core.MapperFunc)

	for _, row := range data {
		log.Printf("fetching block heights between %v and %v for[%v]", row.From, row.To, row.CoinName)
		res := covalent.GetBlockHeights(t.ETH, row.From, row.From.Add(time.Hour))
		startBlock := res.Data.Items[0]
		res = covalent.GetBlockHeights(t.ETH, row.To, row.To.Add(time.Hour))
		endBlock := res.Data.Items[0]

		events := covalent.GetLogEvents(row.ContractAddr, startBlock.Height, endBlock.Height, t.ETH)

		var transferEvents []t.TransferEvent
		for _, item := range events.Data.Items {
			if item.DecodedEvent.Name == "Transfer" {
				from, _ := item.DecodedEvent.Params[0].Value.(string)
				to, _ := item.DecodedEvent.Params[1].Value.(string)
				foo, _ := item.DecodedEvent.Params[2].Value.(string)
				bar, _ := strconv.ParseFloat(foo, 64)

				transferEvents = append(transferEvents, t.TransferEvent{
					FromAddr: from,
					ToAddr:   to,
					Amount:   bar / math.Pow10(item.ContractDecimals),
					CoinName: item.ContractTickerSymbol,
				})
			}
		}
		histories = append(histories, transferEvents)
	}

	for idx, h := range histories {
		md := core.MergeDuplicates(h, data[idx].Rate)
		uniqueHistories = append(uniqueHistories, md)
	}

	crossRef := core.Intersection(uniqueHistories)
	result := core.FilterContracts(crossRef)

	sort.Slice(result, func(i, j int) bool { return result[i].Pumps > result[j].Pumps })

	fname := utils.GenFileName(time.Now())
	core.DownloadCSV(fname, w, core.CSVHeaders, result)
}
