package core

import (
	"strconv"
	"sync"
	"time"

	"github.com/gsrai/go-ionic/internal/clients/etherscan"
	t "github.com/gsrai/go-ionic/internal/types"
	"github.com/gsrai/go-ionic/internal/utils"
)

var CSVHeaders = []string{
	"Wallet address",
	"Number of pumps",
	"Coins traded",
	"Total spent (USD)",
	"Number of trades",
}

func worker(jobs <-chan string, results chan<- string, wg *sync.WaitGroup) {
	for j := range jobs {
		if etherscan.IsContract(j) {
			results <- j
		}
	}
	wg.Done()
}

func FilterContracts(bar map[string]t.WalletPumpHistory) []t.OutputCSVRecord {
	var result []t.OutputCSVRecord
	var wg sync.WaitGroup
	jobs := make(chan string, 5)
	results := make(chan string, 5)
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
		result = append(result, t.OutputCSVRecord{
			Address:  r,
			Trades:   v.Trades,
			Pumps:    v.Pumps,
			SumTotal: v.SumTotal,
			Coins:    v.Coins,
		})
	}
	return result
}

func Intersection(uniqueHistories [][]t.CoinTradeInfo) map[string]t.WalletPumpHistory {
	result := make(map[string]t.WalletPumpHistory)
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
					result[record.Address] = t.WalletPumpHistory{
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
