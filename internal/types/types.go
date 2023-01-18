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

	return []string{a, p, c, s, t}
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
	Trades   int
	SumTotal float64
	Pumps    int
	Coins    map[string]struct{}
}

type CoinTradeInfo struct {
	Address    string
	CoinName   string
	Occurrence int
	SumTotal   float64
}
