package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/gsrai/go-ionic/internal/clients/covalent"
	"github.com/gsrai/go-ionic/internal/clients/etherscan"
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
	http.HandleFunc("/get_wallets", getWallets)

	log.Print("Running server on " + addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func getWallets(w http.ResponseWriter, req *http.Request) {
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

	var histories [][]TransferEvent

	for _, row := range data {
		log.Printf("fetching block heights between %v and %v for[%v]", row.From, row.To, row.CoinName)
		res := covalent.GetBlockHeights(types.ETH, row.From, row.From.Add(time.Hour))
		startBlock := res.Data.Items[0]
		res = covalent.GetBlockHeights(types.ETH, row.To, row.To.Add(time.Hour))
		endBlock := res.Data.Items[0]

		events := covalent.GetLogEvents(types.Address(row.ContractAddr), startBlock.Height, endBlock.Height, types.ETH)

		var transferEvents []TransferEvent
		for _, item := range events.Data.Items {
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
		histories = append(histories, transferEvents)
	}

	var uniqueHistories [][]CoinTradeInfo
	for idx, h := range histories {
		md := mergeDuplicates(h, data[idx].Rate)
		uniqueHistories = append(uniqueHistories, md)
	}
	crossRef := intersection(uniqueHistories)
	result := filterContracts(crossRef)

	sort.Slice(result, func(i, j int) bool { return result[i].Pumps > result[j].Pumps })

	fname := fmt.Sprintf("wallets_%s.csv", time.Now().Format("2006-01-02_15:04:05"))
	csvHeaders := []string{
		"Wallet address",
		"Number of pumps",
		"Coins traded",
		"Total spent (USD)",
		"Number of trades",
	}
	csv.Download(fname, w, csvHeaders, result)
}

func worker(jobs <-chan types.Address, results chan<- types.Address, wg *sync.WaitGroup) {
	for j := range jobs {
		if etherscan.IsContract(j) {
			results <- j
		}
	}
	wg.Done()
}

func filterContracts(bar map[types.Address]WalletPumpHistory) []types.OutputCSVRecord {
	var result []types.OutputCSVRecord
	var wg sync.WaitGroup
	jobs := make(chan types.Address, 5)
	results := make(chan types.Address, 5)
	wg.Add(5)
	go worker(jobs, results, &wg)
	go worker(jobs, results, &wg)
	go worker(jobs, results, &wg)
	go worker(jobs, results, &wg)
	go worker(jobs, results, &wg)

	go func() {
		counter := 0
		for addr := range bar {
			if counter%5 == 0 {
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

	for r := range results {
		v := bar[r]
		result = append(result, types.OutputCSVRecord{
			Address:  r,
			Trades:   v.Trades,
			Pumps:    v.Pumps,
			SumTotal: v.SumTotal,
			Coins:    v.Coins,
		})
	}
	return result
}

type WalletPumpHistory struct {
	Trades   int
	SumTotal float64
	Pumps    int
	Coins    []string
}

func intersection(uniqueHistories [][]CoinTradeInfo) map[types.Address]WalletPumpHistory {
	result := make(map[types.Address]WalletPumpHistory)
	for i := 0; i < len(uniqueHistories); i++ {
		for j := 0; j < len(uniqueHistories[i]); j++ {
			record := uniqueHistories[i][j]
			for k := i + 1; k < len(uniqueHistories); k++ {
				index := -1
				for idx, ele := range uniqueHistories[k] {
					if ele.Address == record.Address {
						index = idx
						break
					}
				}
				if index == -1 {
					continue
				}
				if wallet, pres := result[record.Address]; pres {
					wallet.Trades += uniqueHistories[k][index].Occurrence
					wallet.SumTotal += uniqueHistories[k][index].SumTotal
					wallet.Pumps++
					wallet.Coins = append(wallet.Coins, uniqueHistories[k][index].CoinName)
					result[record.Address] = wallet
				} else {
					result[record.Address] = WalletPumpHistory{
						Trades:   uniqueHistories[k][index].Occurrence + record.Occurrence,
						SumTotal: uniqueHistories[k][index].SumTotal + record.SumTotal,
						Pumps:    2,
						Coins:    []string{uniqueHistories[k][index].CoinName, record.CoinName},
					}
				}
			}
		}
	}
	return result
}

type CoinTradeInfo struct {
	Address    types.Address
	CoinName   string
	Occurrence int
	SumTotal   float64
}

func mergeDuplicates(eventLog []TransferEvent, rate float64) []CoinTradeInfo {
	var sli []CoinTradeInfo
	m := make(map[types.Address]CoinTradeInfo)
	for _, entry := range eventLog {
		if cti, pres := m[entry.To]; pres {
			cti.Occurrence += 1
			cti.SumTotal += entry.Amount
			m[entry.To] = cti
		} else {
			m[entry.To] = CoinTradeInfo{
				Address:    entry.To,
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
