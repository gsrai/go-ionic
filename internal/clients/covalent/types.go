package covalent

import "time"

type APIResponse interface {
	Block | LogEvent
}

type Block struct {
	SignedAt time.Time `json:"signed_at"`
	Height   int       `json:"height"`
}
type LogEvent struct {
	DecodedEvent struct {
		Name   string `json:"name"`
		Params []struct {
			Value interface{} `json:"value"`
		} `json:"params"`
	} `json:"decoded"`
	ContractDecimals     int    `json:"sender_contract_decimals"`
	ContractTickerSymbol string `json:"sender_contract_ticker_symbol"`
}

type CovalentPagination struct {
	HasMore    bool `json:"has_more"`
	PageNumber int  `json:"page_number"`
	PageSize   int  `json:"page_size"`
}

type CovalentDataBody[T APIResponse] struct {
	UpdatedAt  time.Time          `json:"updated_at"`
	Items      []T                `json:"items"`
	Pagination CovalentPagination `json:"pagination"`
}

type CovalentAPIResponse[T APIResponse] struct {
	Data         CovalentDataBody[T] `json:"data"`
	Error        bool                `json:"error"`
	ErrorMessage string              `json:"error_message"`
	ErrorCode    int                 `json:"error_code"`
}
