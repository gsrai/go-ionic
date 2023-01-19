package types

import (
	"fmt"
	"strconv"
	"time"
)

type InputCSVRecord struct {
	CoinName     string
	ContractAddr string
	From         time.Time
	To           time.Time
	Network      string
	Rate         float64
}

type OutputCSVRecord struct {
	Address  string
	Trades   int
	SumTotal float64
	Pumps    int
	Coins    []string
}

func (o OutputCSVRecord) ToSlice() []string {
	a := o.Address
	p := strconv.Itoa(o.Pumps)
	c := fmt.Sprintf("%v", o.Coins)
	t := strconv.Itoa(o.Trades)
	s := fmt.Sprintf("%.2f", o.SumTotal)

	return []string{a, p, c, t, s}
}

type ChainID int

const (
	ETH ChainID = iota + 1
	BSC
)

type TransferEvent struct {
	FromAddr string
	ToAddr   string
	Amount   float64
	CoinName string
}

type WalletPumpHistory struct {
	Address  string
	Trades   int
	SumTotal float64
	Pumps    int
	Coins    map[string]struct{}
}

func NewWalletPumpHistory(address string) *WalletPumpHistory {
	return &WalletPumpHistory{Address: address, Coins: map[string]struct{}{}}
}

func (w *WalletPumpHistory) AddTransfer(coinName string, amount, rate float64) {
	if _, pres := w.Coins[coinName]; !pres {
		w.Coins[coinName] = struct{}{}
		w.Pumps++
	}
	w.Trades++
	w.SumTotal += amount * rate
}

type CoinTradeInfo struct {
	Address    string
	CoinName   string
	Occurrence int
	SumTotal   float64
}
