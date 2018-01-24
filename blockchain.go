package minichain

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"os"
	"time"
)

type BlockChain struct {
	file   *os.File
	ticker *time.Ticker

	BlockSize int
	LastBlock *Block
	Timeout   time.Duration
	Input     chan *Transaction
	ShutDown  chan chan struct{}
}

func NewBlockChain(config *Config) (*BlockChain, error) {
	file, err := os.OpenFile(config.BlockChain.DataFile, os.O_SYNC|os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)

	if err != nil {
		return nil, err
	}

	m := &BlockChain{
		file:      file,
		ticker:    time.NewTicker(time.Second * time.Duration(config.BlockChain.TimeOut)),
		BlockSize: config.BlockChain.BlockSize,
		Timeout:   time.Duration(config.BlockChain.TimeOut) * time.Second,
		Input:     make(chan *Transaction),
		ShutDown:  make(chan chan struct{}),
	}

	go m.Run()

	return m, err
}

func (b BlockChain) Run() {
	var transactions = make([]*Transaction, 0, b.BlockSize)

	for {
		select {
		case ch := <-b.ShutDown:
			GetLogger().Info("Shutdown blockchain")
			if err := b.Flush(transactions); err != nil {
				GetLogger().Error(err)
			}

			if err := b.file.Close();err != nil {
				GetLogger().Error(err)
			}

			close(ch)
			return
		case tx := <-b.Input:
			GetLogger().Infof("Receive transaction %v", tx)
			transactions = append(transactions, tx)

			if len(transactions) == b.BlockSize {
				if err := b.Flush(transactions); err != nil {
					GetLogger().Error(err)
				}
				// Reset ticket after transaction pool overflow
				b.ticker = time.NewTicker(b.Timeout)
				transactions = make([]*Transaction, 0, b.BlockSize)
			}
		case <-b.ticker.C:
			GetLogger().Info("Flush by ticker")
			if err := b.Flush(transactions); err != nil {
				GetLogger().Error(err)
			}

			transactions = make([]*Transaction, 0, b.BlockSize)
		}
	}
}

func (b BlockChain) Flush(transactions []*Transaction) error {
	var block *Block

	if len(transactions) == 0 {
		return nil
	}

	if b.LastBlock != nil {
		block = NewBlock(b.LastBlock.BlockHash, transactions)
	} else {
		block = NewBlock([]byte("Origin"), transactions)
	}

	blockBytes, err := json.Marshal(block)

	if err != nil {
		return err
	}

	blockSize := uint32(len(blockBytes))
	header := make([]byte, 4)
	binary.LittleEndian.PutUint32(header, blockSize)

	data := bytes.Join([][]byte{header, blockBytes}, []byte{})
	n, err := b.file.Write(data)

	if err != nil {
		return err
	}

	GetLogger().Debugf("Bytes written %d", n)
	b.file.Sync()

	if err != nil {
		return err
	}

	return nil
}
