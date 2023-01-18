package core

import (
	"encoding/csv"
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

const API_LIMIT = 5 // rate limit in requests per second

var CSVHeaders = []string{
	"Wallet address",
	"Number of pumps",
	"Coins traded",
	"Number of trades",
	"Total spent (USD)",
}

func processRow(row t.InputCSVRecord) []t.TransferEvent {
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
	return transferEvents
}

func transfersWorker(jobs <-chan t.InputCSVRecord, results chan<- []t.TransferEvent, wg *sync.WaitGroup) {
	for j := range jobs {
		results <- processRow(j)
	}
	wg.Done()
}

func GetTransferEvents(data []t.InputCSVRecord) [][]t.TransferEvent {
	var transfers [][]t.TransferEvent
	var wg sync.WaitGroup
	jobs := make(chan t.InputCSVRecord, API_LIMIT)
	results := make(chan []t.TransferEvent, API_LIMIT)
	wg.Add(API_LIMIT)

	for i := 0; i < API_LIMIT; i++ {
		go transfersWorker(jobs, results, &wg)
	}

	go func() {
		counter := 0
		for _, row := range data {
			if counter%API_LIMIT == 0 {
				time.Sleep(time.Second)
				counter = 0
			}
			jobs <- row
			counter++
		}
		close(jobs)
		wg.Wait()
		close(results)
	}()

	for eventlog := range results {
		transfers = append(transfers, eventlog)
	}
	return transfers
}

func worker(jobs <-chan string, results chan<- string, wg *sync.WaitGroup) {
	for j := range jobs {
		if etherscan.IsContract(j) {
			results <- j
		}
	}
	wg.Done()
}

func FilterContracts(wallets map[string]t.WalletPumpHistory) []t.OutputCSVRecord {
	var filteredWallets []t.OutputCSVRecord
	var wg sync.WaitGroup
	jobs := make(chan string, API_LIMIT)
	results := make(chan string, API_LIMIT)
	wg.Add(API_LIMIT)

	for i := 0; i < API_LIMIT; i++ {
		go worker(jobs, results, &wg)
	}

	log.Printf("Total wallets: %v", len(wallets))
	go func() {
		counter := 0
		for addr := range wallets {
			if counter%API_LIMIT == 0 {
				time.Sleep(time.Second)
				counter = 0
			}
			jobs <- addr
			counter++
		}
		close(jobs)
		wg.Wait()
		close(results)
	}()

	for addr := range results {
		v := wallets[addr]
		c := utils.GetMapKeys(v.Coins)
		filteredWallets = append(filteredWallets, t.OutputCSVRecord{
			Address:  addr,
			Trades:   v.Trades,
			Pumps:    v.Pumps,
			SumTotal: v.SumTotal,
			Coins:    c,
		})
	}
	return filteredWallets
}

// this was bugged, but could also be optimized:
// we can just concatinate the eventlogs and go through them once, adding to the map.
// currently we are going through each coin's eventlog and comparing it to the other coins' eventlogs
// this has a worst case complexity of O(n^3)
// concatination would be O(n) and then we can go through the eventlog once, adding to the map.
func Intersection(uniqueHistories [][]t.CoinTradeInfo) map[string]t.WalletPumpHistory {
	result := make(map[string]t.WalletPumpHistory)
	// each coin's deduped eventlog (list of transfer events)
	for i := 0; i < len(uniqueHistories); i++ {
		// each item is an wallet address in the deduped event log
		for j := 0; j < len(uniqueHistories[i]); j++ {
			record := uniqueHistories[i][j]
			// skip if address is in result map
			if _, pres := result[record.Address]; pres {
				continue
			}
			for k := i + 1; k < len(uniqueHistories); k++ {
				index := -1
				for idx, ele := range uniqueHistories[k] {
					if ele.Address == record.Address {
						index = idx // only one address per eventlog as it was deduped
						break
					}
				}
				if index == -1 {
					continue
				}
				// if found a duplicate address, find and add in result map
				if wallet, pres := result[record.Address]; pres {
					wallet.Trades += uniqueHistories[k][index].Occurrence
					wallet.SumTotal += uniqueHistories[k][index].SumTotal
					wallet.Pumps++
					wallet.Coins[uniqueHistories[k][index].CoinName] = struct{}{}
					result[record.Address] = wallet
				} else {
					result[record.Address] = t.WalletPumpHistory{
						Trades:   uniqueHistories[k][index].Occurrence + record.Occurrence,
						SumTotal: uniqueHistories[k][index].SumTotal + record.SumTotal,
						Pumps:    2,
						Coins: map[string]struct{}{
							record.CoinName:                    {},
							uniqueHistories[k][index].CoinName: {},
						},
					}
				}
			}
		}
	}
	return result
}

func CollateData(data [][]t.CoinTradeInfo) map[string]t.WalletPumpHistory {
	result := make(map[string]t.WalletPumpHistory)
	for _, coin := range data {
		for _, entry := range coin {
			if wallet, pres := result[entry.Address]; pres {
				wallet.Trades += entry.Occurrence
				wallet.SumTotal += entry.SumTotal
				wallet.Pumps++
				wallet.Coins[entry.CoinName] = struct{}{}
				result[entry.Address] = wallet
			} else {
				result[entry.Address] = t.WalletPumpHistory{
					Trades:   entry.Occurrence,
					SumTotal: entry.SumTotal,
					Pumps:    1,
					Coins: map[string]struct{}{
						entry.CoinName: {},
					},
				}
			}
		}
	}
	return result
}

func MergeDuplicates(eventLog []t.TransferEvent, rate float64) []t.CoinTradeInfo {
	var sli []t.CoinTradeInfo
	m := make(map[string]t.CoinTradeInfo)
	for _, entry := range eventLog {
		if cti, pres := m[entry.ToAddr]; pres {
			cti.Occurrence += 1
			cti.SumTotal += entry.Amount
			m[entry.ToAddr] = cti
		} else {
			m[entry.ToAddr] = t.CoinTradeInfo{
				Address:    entry.ToAddr,
				Occurrence: 1,
				SumTotal:   entry.Amount,
				CoinName:   entry.CoinName,
			}
		}
	}

	for _, v := range m {
		v.SumTotal *= rate
		sli = append(sli, v)
	}
	return sli
}

func DownloadCSV(fileName string, w http.ResponseWriter, headers []string, content []t.OutputCSVRecord) {
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
