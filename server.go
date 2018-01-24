package minichain

import "net/http"

type BlockChainServer struct {
	MemPool *MemPool
}

func NewBlockChainServer(config *Config) *BlockChainServer {
	return &BlockChainServer{
		NewMemPool(config.BlockChain.BlockSize,
			config.BlockChain.TimeOut),
	}
}

func (blockChainServer *BlockChainServer) TransactionHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")

	if len(key) == 0 {
		http.Error(w, "Key cannot be empty", http.StatusBadRequest)
	}

	value := r.URL.Query().Get("value")

	tx := NewTransaction([]byte(key), []byte(value))
	blockChainServer.MemPool.Input <- tx
}
