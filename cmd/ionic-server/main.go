package main

import (
	"fmt"
	"github.com/gsrai/go-ionic/internal/misc"
	"log"
	"net/http"
	"regexp"
	"sort"
	"sync"
	"time"

	"github.com/gsrai/go-ionic/config"
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
	http.HandleFunc("/v2/", getPumps)

	log.Print("Running server on " + addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func getWallets(w http.ResponseWriter, req *http.Request) {
	defer func(t time.Time) {
		elapsed := time.Since(t)
		fmt.Printf("finished getting wallets in %s 👍\n", elapsed)
	}(time.Now())

	var uniqueHistories [][]t.CoinTradeInfo

	data := csv.ReadAndParse(config.Get().InputFilePath, core.MapperFunc)
	transfers := core.GetTransferEvents(data)

	for idx, h := range transfers {
		md := core.MergeDuplicates(h, data[idx].Rate)
		uniqueHistories = append(uniqueHistories, md)
	}

	// crossRef := core.Intersection(uniqueHistories)
	crossRef := core.CollateData(uniqueHistories)

	// filter out contracts with less than 3 pumps
	var filtered map[string]t.WalletPumpHistory = make(map[string]t.WalletPumpHistory)
	for addr, item := range crossRef {
		if item.Pumps >= 3 {
			filtered[addr] = item
		}
	}

	result := core.FilterContracts(filtered)

	sort.Slice(result, func(i, j int) bool { return result[i].Pumps > result[j].Pumps })

	fname := utils.GenFileName(time.Now())
	core.DownloadCSV(fname, w, core.CSVHeaders, result)
}

// ---------

func getPumps(w http.ResponseWriter, _ *http.Request) {
	defer func(t time.Time) {
		elapsed := time.Since(t)
		fmt.Printf("finished getting wallets in %s 👍\n", elapsed)
	}(time.Now())

	r := regexp.MustCompile(`\(([^\)]+)\)`)
	parseCoinName := func(s string) string {
		res := r.FindString(s)
		return res[1 : len(res)-1]
	}

	coinRates := make(map[string]float64)

	dataIngest := make(chan t.InputCSVRecord, 5)
	fooChan := make(chan t.InputCSVRecord, 5) // TODO rename!
	blockHeights := make(chan misc.LogEventQuery, 5)
	transferEvents := make(chan []t.TransferEvent, 5)
	eventLog := make(chan t.TransferEvent, 500)
	wallets := make(chan t.WalletPumpHistory, 500)
	dataOutStream := make(chan t.OutputCSVRecord, 500)
	go csv.ReadAndParseP(config.Get().InputFilePath, core.MapperFunc, dataIngest)
	go func(in <-chan t.InputCSVRecord, out chan<- t.InputCSVRecord) {
		defer close(out)
		for coin := range in {
			c := parseCoinName(coin.CoinName)
			coinRates[c] = coin.Rate
			out <- coin
		}
	}(dataIngest, fooChan)
	go core.GetBlockHeights(fooChan, blockHeights)
	go core.FetchLogEvents(blockHeights, transferEvents)

	go func(in <-chan []t.TransferEvent, out chan<- t.TransferEvent) {
		log.Println("unpacking logevents")
		defer close(out)
		for el := range in { // el => eventlog
			for _, event := range el {
				out <- event
			}
		}
		log.Println("unpacked logevents")
	}(transferEvents, eventLog)

	go core.CollateDataP(eventLog, wallets, coinRates)
	// don't forget to wait!
	var wg sync.WaitGroup
	go func(in <-chan t.WalletPumpHistory, out chan<- t.OutputCSVRecord) {
		defer close(out)
		for wallet := range in {

			c := utils.GetMapKeys(wallet.Coins)
			out <- t.OutputCSVRecord{
				Address:  wallet.Address,
				Trades:   wallet.Trades,
				Pumps:    wallet.Pumps,
				SumTotal: wallet.SumTotal,
				Coins:    c,
			}
		}
		log.Printf("streaming csv data\n")
	}(wallets, dataOutStream)
	wg.Add(1)
	core.DownloadCSVP(w, core.CSVHeaders, dataOutStream) // why not use core.CSVHeaders directly?
	wg.Done()
}
