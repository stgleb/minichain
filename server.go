package minichain

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type BlockChainServer struct {
	KeyMaxSize   int
	ValueMaxSize int
	Timeout      time.Duration
	BlockChain   *BlockChain
}

func NewBlockChainServer(config *Config) (*BlockChainServer, error) {
	blockChain, err := NewBlockChain(config)

	if err != nil {
		return nil, err
	}

	return &BlockChainServer{
		KeyMaxSize:   config.BlockChain.KeyMaxSize,
		ValueMaxSize: config.BlockChain.ValueMaxSize,
		Timeout:      time.Duration(config.Http.Timeout) * time.Second,
		BlockChain:   blockChain,
	}, nil
}

func (blockChainServer *BlockChainServer) TransactionHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")

	if len(key) == 0 {
		http.Error(w, "Key cannot be empty", http.StatusBadRequest)
	}

	if len(key) > blockChainServer.KeyMaxSize {
		http.Error(w, fmt.Sprintf("Key size is too long %d max allowed %d",
			len(key), blockChainServer.KeyMaxSize),
			http.StatusBadRequest)
		return
	}

	value := r.URL.Query().Get("value")

	if len(key) > blockChainServer.ValueMaxSize {
		http.Error(w, fmt.Sprintf("Value size is too long %d max allowed %d",
			len(value), blockChainServer.ValueMaxSize),
			http.StatusBadRequest)
		return
	}

	GetLogger().Infof("Create new transaction key %s value %s", key, value)

	tx := NewTransaction(key, value)
	blockChainServer.BlockChain.Input <- tx

	// Status is accepted since transaction flushes to disk asynchronously
	w.WriteHeader(http.StatusAccepted)
}

func (blockChainServer *BlockChainServer) SearchByKey(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")

	if len(key) == 0 {
		http.Error(w, "Key cannot be empty", http.StatusBadRequest)
	}

	resultChan := make(chan *SearchResult)
	ctx, cancel := context.WithTimeout(r.Context(), blockChainServer.Timeout)
	defer cancel()

	req := &SearchRequest{
		ctx,
		key,
		resultChan,
	}

	blockChainServer.BlockChain.Search <- req

	select {
	case <-ctx.Done():
		// Lets consider timeout for search as a timeout of requesting another service
		http.Error(w, "search request timed out", http.StatusGatewayTimeout)
	case searchResult := <-resultChan:
		if len(searchResult.Transactions) == 0 {
			w.WriteHeader(http.StatusNotFound)
		}

		json.NewEncoder(w).Encode(searchResult)
	}
}
