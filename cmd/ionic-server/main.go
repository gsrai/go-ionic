package main

import (
	"fmt"
	"log"
	"net/http"
	"sort"
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
