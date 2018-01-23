package minichain

import "net/http"

type BlockChainServer struct {
}

func (blockChainServer *BlockChainServer) TransactionHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")

	if len(key) == 0 {
		http.Error(w, "Key cannot be empty", http.StatusBadRequest)
	}

	value := r.URL.Query().Get("value")

	_, err := NewTransaction(key, value)

	if err != nil {
		http.Error(w, "Error while creating transaction", http.StatusInternalServerError)
	}
}
