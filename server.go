package minichain

import "net/http"

type BlockChainServer struct {
	memPool *MemPool
}

func NewBlockChainServer() *BlockChainServer {
	return &BlockChainServer{
		NewMemPool(1),
	}
}

func (blockChainServer *BlockChainServer) TransactionHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")

	if len(key) == 0 {
		http.Error(w, "Key cannot be empty", http.StatusBadRequest)
	}

	value := r.URL.Query().Get("value")

	tx := NewTransaction([]byte(key), []byte(value))
	blockChainServer.memPool.Input <- tx
}
