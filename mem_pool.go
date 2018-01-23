package minichain

import (
	"os"
	"time"
)

type MemPool struct {
	file   *os.File
	ticker *time.Ticker

	Input    chan Transaction
	ShutDown chan struct{}
}

func NewMemPool(period int) *MemPool {
	m := &MemPool{
		ticker:   time.NewTicker(time.Second * time.Duration(period)),
		Input:    make(chan Transaction),
		ShutDown: make(chan struct{}),
	}

	m.Run()

	return m
}

func (m MemPool) Run() {
	var b *Block

	for {
		select {
		case <-m.ShutDown:
			m.Flush(b)
			return
		case tx := <-m.Input:
			GetLogger().Infof("Receive transaction %v", tx)
			b.Transactions = append(b.Transactions, tx)
		case <-m.ticker.C:
			m.Flush(b)
		}
	}
}

func (m MemPool) Flush(block *Block) {}
