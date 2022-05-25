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
	Address  Address
	Trades   int
	SumTotal float64
	Pumps    int
	Coins    []string
}

func (o OutputCSVRecord) ToSlice() []string {
	a := string(o.Address)
	p := strconv.Itoa(o.Pumps)
	c := fmt.Sprintf("%q", o.Coins)
	s := fmt.Sprintf("%.2f", o.SumTotal)
	t := strconv.Itoa(o.Trades)

	return []string{a, p, c, s, t}
}

type ChainID int

const (
	ETH ChainID = iota + 1
	BSC
)

type Address string
