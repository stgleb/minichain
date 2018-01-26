package minichain

import "context"

type SearchRequest struct {
	ctx        context.Context
	Key        string
	ResultChan chan *SearchResult
}

type SearchResult struct {
	Transactions []Transaction `json:"transaction"`
	Error        string        `json:"error"`
}
