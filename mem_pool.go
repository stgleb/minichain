package minichain

import (
	"bytes"
	"encoding/json"
	"os"
	"strconv"
	"time"
)

type MemPool struct {
	file   *os.File
	ticker *time.Ticker

	LastBlock *Block
	Input     chan *Transaction
	ShutDown  chan struct{}
}

func NewMemPool(period int) *MemPool {
	m := &MemPool{
		ticker:   time.NewTicker(time.Second * time.Duration(period)),
		Input:    make(chan *Transaction),
		ShutDown: make(chan struct{}),
	}

	go m.Run()

	return m
}

func (m MemPool) Run() {
	var transactions = make([]*Transaction, 0, 10)

	for {
		select {
		case <-m.ShutDown:
			m.Flush(transactions)
			return
		case tx := <-m.Input:
			GetLogger().Infof("Receive transaction %v", tx)
			transactions = append(transactions, tx)
		case <-m.ticker.C:
			m.Flush(transactions)
		}
	}
}

func (m MemPool) Flush(transactions []*Transaction) error {
	var block *Block

	if m.LastBlock != nil {
		block = NewBlock(m.LastBlock.BlockHash, transactions)
	} else {
		block = NewBlock([]byte("Origin"), transactions)
	}

	txCount := len(block.Transactions)

	blockBytes, err := json.Marshal(block)
	header := []byte(strconv.Itoa(txCount))

	if err != nil {
		return err
	}

	data := bytes.Join([][]byte{header, blockBytes}, []byte{})
	_, err = m.file.Write(data)

	if err != nil {
		return err
	}

	return nil
}
