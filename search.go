package minichain

import "context"

type SearchRequest struct {
	ctx        context.Context
	Key        string
	ResultChan chan *SearchResult
}

type SearchResult struct {
	Transaction *Transaction `json:"transaction"`
	Error       error        `json:"error"`
}
