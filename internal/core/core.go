package core

import (
	"encoding/csv"
	"github.com/gsrai/go-ionic/internal/misc"
	"log"
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gsrai/go-ionic/internal/clients/covalent"
	"github.com/gsrai/go-ionic/internal/clients/etherscan"
	t "github.com/gsrai/go-ionic/internal/types"
	"github.com/gsrai/go-ionic/internal/utils"
)

const ApiLimit = 5 // rate limit in requests per second

var CSVHeaders = []string{
	"Wallet address",
	"Number of pumps",
	"Coins traded",
	"Number of trades",
	"Total spent (USD)",
}

func worker[I any, O any](jobs <-chan I, results chan<- O, wg *sync.WaitGroup, fn func(I) O) {
	for j := range jobs {
		results <- fn(j) // what if this is a conditional or a slice?
	}
	wg.Done()
}

func GetBlockHeights(in <-chan t.InputCSVRecord, out chan<- misc.LogEventQuery) {
	defer close(out)
	var wg sync.WaitGroup
	jobs := make(chan t.InputCSVRecord, ApiLimit)
	for i := 0; i < ApiLimit; i++ {
		go worker(jobs, out, &wg, func(record t.InputCSVRecord) misc.LogEventQuery {
			log.Printf("fetching block heights between %v and %v for[%v]", record.From, record.To, record.CoinName)
			res := covalent.GetBlockHeights(t.ETH, record.From, record.From.Add(time.Hour))
			startBlock := res.Data.Items[0]
			res = covalent.GetBlockHeights(t.ETH, record.To, record.To.Add(time.Hour))
			endBlock := res.Data.Items[0]
			return misc.LogEventQuery{
				InputCSVRecord: record,
				StartBlock:     startBlock,
				EndBlock:       endBlock,
			}
		})
		wg.Add(1)
	}
	counter := 0
	for coin := range in {
		if counter%ApiLimit == 0 {
			time.Sleep(time.Second)
			counter = 0
		}
		jobs <- coin
		counter++
	}
	close(jobs)
	wg.Wait()
}

func FetchLogEvents(in <-chan misc.LogEventQuery, out chan<- []t.TransferEvent) {
	defer close(out)
	var wg sync.WaitGroup
	jobs := make(chan misc.LogEventQuery, ApiLimit)
	for i := 0; i < ApiLimit; i++ {
		go worker(jobs, out, &wg, func(record misc.LogEventQuery) []t.TransferEvent {
			events := covalent.GetLogEvents(record.ContractAddr, record.StartBlock.Height, record.EndBlock.Height, t.ETH)

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
			return transferEvents
		})
		wg.Add(1)
	}
	counter := 0
	for coin := range in {
		if counter%ApiLimit == 0 {
			time.Sleep(time.Second) // this can be improved
			counter = 0
		}
		jobs <- coin
		counter++
	}
	close(jobs)
	wg.Wait()
}

func CollateData(in <-chan t.TransferEvent, out chan<- t.WalletPumpHistory, coinRates map[string]float64) {
	result := make(map[string]*t.WalletPumpHistory)
	for event := range in {
		rate := coinRates[event.CoinName]
		if _, pres := result[event.ToAddr]; !pres {
			result[event.ToAddr] = t.NewWalletPumpHistory(event.ToAddr)
		}
		result[event.ToAddr].AddTransfer(event.CoinName, event.Amount, rate)
	}

	log.Printf("collated data, size: %v\n", len(result))

	wallets := make(chan t.WalletPumpHistory, 500)
	go func() {
		defer close(wallets)
		for _, wallet := range result {
			if wallet.Pumps > 2 {
				wallets <- *wallet
			}
		}
	}()

	FilterContracts(wallets, out)
}

func worker3(jobs <-chan t.WalletPumpHistory, results chan<- t.WalletPumpHistory, wg *sync.WaitGroup) {
	defer wg.Done()
	for j := range jobs {
		if !etherscan.IsContract(j.Address) {
			results <- j
		}
	}
}

func FilterContracts(in <-chan t.WalletPumpHistory, out chan<- t.WalletPumpHistory) {
	defer close(out)
	// TODO: weird bug, if i have more than 2 go routines work on this, i get inconsistent results
	var wg sync.WaitGroup
	jobs := make(chan t.WalletPumpHistory, 2)
	for i := 0; i < 2; i++ {
		go worker3(jobs, out, &wg)
		wg.Add(1)
	}

	counter := 0
	for wallet := range in {
		if counter%2 == 0 {
			time.Sleep(time.Second)
			counter = 0
		}
		jobs <- wallet
		counter++
	}
	close(jobs)
	wg.Wait()
}

func DownloadCSV(w http.ResponseWriter, headers []string, content <-chan t.OutputCSVRecord) {
	filename := utils.GenFileName(time.Now())

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment;filename="+filename)
	w.Header().Set("Transfer-Encoding", "chunked")
	writer := csv.NewWriter(w)
	err := writer.Write(headers)
	if err != nil {
		http.Error(w, "Error sending csv: "+err.Error(), http.StatusInternalServerError)
		return
	}
	count := 0
	for row := range content {
		count++
		ss := row.ToSlice()
		err := writer.Write(ss)
		if err != nil {
			http.Error(w, "Error sending csv: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
	writer.Flush()
	log.Println(">>> ", count)
}

func MapperFunc(csvRow []string) (t.InputCSVRecord, error) {
	rate, err := strconv.ParseFloat(csvRow[5], 64)
	if err != nil {
		return t.InputCSVRecord{}, err
	}

	from, err := utils.ParseDateTime(csvRow[3])
	if err != nil {
		return t.InputCSVRecord{}, err
	}

	to, err := utils.ParseDateTime(csvRow[4])
	if err != nil {
		return t.InputCSVRecord{}, err
	}

	return t.InputCSVRecord{
		CoinName:     csvRow[0],
		ContractAddr: csvRow[1],
		From:         from,
		To:           to,
		Network:      csvRow[2],
		Rate:         rate,
	}, nil
}
