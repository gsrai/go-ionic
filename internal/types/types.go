package types

import "time"

type InputCSVRecord struct {
	CoinName     string
	ContractAddr string
	From         time.Time
	To           time.Time
	Network      string
	Rate         float64
}
