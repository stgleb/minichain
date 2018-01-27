package minichain

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestBlockChainServerTransactionHandler(t *testing.T) {
	testData := []struct {
		Key          string
		Value        string
		ExpectedCode int
	}{
		{
			Key:          "hello",
			Value:        "world",
			ExpectedCode: http.StatusAccepted,
		},
		{
			Key:          "toolongkey",
			Value:        "world",
			ExpectedCode: http.StatusBadRequest,
		},
		{
			Key:          "hello",
			Value:        "toolongvalue",
			ExpectedCode: http.StatusBadRequest,
		},
	}

	for _, test := range testData {
		blockChain := &BlockChain{
			Input: make(chan *Transaction, 1),
		}

		blockChainServer := &BlockChainServer{
			5,
			5,
			time.Second,
			blockChain,
		}

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/tx?key=%s&value=%s",
			test.Key, test.Value), nil)
		blockChainServer.TransactionHandler(w, req)

		if w.Code != test.ExpectedCode {
			t.Errorf("Wrong response code expected %d actual %d", test.ExpectedCode, w.Code)
		}
	}
}

func TestNewBlockChainServerSearch(t *testing.T) {
	blockChain := &BlockChain{
		Search: make(chan *SearchRequest, 1),
	}

	blockChainServer := &BlockChainServer{
		5,
		5,
		time.Millisecond,
		blockChain,
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/search?key=key", nil)

	blockChainServer.SearchByKey(w, req)

	if w.Code != http.StatusGatewayTimeout {
		t.Errorf("Wrong response code expected %d actual %d",
			http.StatusGatewayTimeout, w.Code)
	}
}
