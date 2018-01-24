package minichain

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"os"
	"time"
)

type MemPool struct {
	file   *os.File
	ticker *time.Ticker

	BlockSize int
	LastBlock *Block
	Input     chan *Transaction
	ShutDown  chan struct{}
}

func NewMemPool(blockSize, period int) *MemPool {
	m := &MemPool{
		BlockSize: blockSize,
		ticker:    time.NewTicker(time.Second * time.Duration(period)),
		Input:     make(chan *Transaction),
		ShutDown:  make(chan struct{}),
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

			if len(transactions) == m.BlockSize {
				m.Flush(transactions)
			}
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

	txCount := uint32(len(block.Transactions))
	header := make([]byte, 4)
	binary.LittleEndian.PutUint32(header, txCount)

	blockBytes, err := json.Marshal(block)

	if err != nil {
		return err
	}

	data := bytes.Join([][]byte{header, blockBytes}, []byte{})
	_, err = m.file.Write(data)
	m.file.Sync()

	if err != nil {
		return err
	}

	return nil
}
