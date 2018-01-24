package minichain

import "net/http"

type BlockChainServer struct {
	BlockChain *BlockChain
}

func NewBlockChainServer(config *Config) (*BlockChainServer, error) {
	blockChain, err := NewBlockChain(config)

	if err != nil {
		return nil, err
	}

	return &BlockChainServer{
		BlockChain: blockChain,
	}, nil
}

func (blockChainServer *BlockChainServer) TransactionHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")

	if len(key) == 0 {
		http.Error(w, "Key cannot be empty", http.StatusBadRequest)
	}

	value := r.URL.Query().Get("value")

	tx := NewTransaction([]byte(key), []byte(value))
	blockChainServer.BlockChain.Input <- tx
}
