package covalent

import "time"

type Block struct {
	SignedAt time.Time `json:"signed_at"`
	Height   int
}

type BlockHeights struct {
	UpdatedAt  time.Time          `json:"updated_at"`
	Items      []Block            `json:"items"`
	Pagination CovalentPagination `json:"pagination"`
}

type CovalentPagination struct {
	HasMore    bool `json:"has_more"`
	PageNumber int  `json:"page_number"`
	PageSize   int  `json:"page_size"`
}

type CovalentAPIResponse[T any] struct {
	Data         T      `json:"data"`
	Error        bool   `json:"error"`
	ErrorMessage string `json:"error_message"`
	ErrorCode    int    `json:"error_code"`
}
