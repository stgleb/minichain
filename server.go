package minichain

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type BlockChainServer struct {
	Timeout    time.Duration
	BlockChain *BlockChain
}

func NewBlockChainServer(config *Config) (*BlockChainServer, error) {
	blockChain, err := NewBlockChain(config)

	if err != nil {
		return nil, err
	}

	return &BlockChainServer{
		Timeout:    time.Duration(config.Http.Timeout) * time.Second,
		BlockChain: blockChain,
	}, nil
}

func (blockChainServer *BlockChainServer) TransactionHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")

	if len(key) == 0 {
		http.Error(w, "Key cannot be empty", http.StatusBadRequest)
	}

	value := r.URL.Query().Get("value")
	GetLogger().Infof("Create new transaction key %s value %s", key, value)

	tx := NewTransaction(key, value)
	blockChainServer.BlockChain.Input <- tx
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
		http.Error(w, "search request timed out", http.StatusNotFound)
	case searchResult := <-resultChan:
		json.NewEncoder(w).Encode(searchResult)
	}
}
