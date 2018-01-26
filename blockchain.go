package minichain

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"os"
	"time"
)

const (
	HEADER_SIZE   = 4
	DIGEST_SIZE   = 32
	GENESIS_BLOCK = "Genesis block"
)

type BlockChain struct {
	// Current data file descriptor
	file   *os.File
	ticker *time.Ticker

	IndexOn       bool
	DataFileName  string
	Offset        int64
	Index         *InvertedIndex
	BlockSize     int
	LastBlockHash []byte
	Timeout       time.Duration
	Input         chan *Transaction
	ShutDown      chan chan struct{}
	Search        chan *SearchRequest
}

func NewBlockChain(config *Config) (*BlockChain, error) {
	prevBlockHash, err := GetLastBlockHash(config.BlockChain.DataFile)

	if err != nil {
		hash := sha256.Sum256([]byte(GENESIS_BLOCK))
		prevBlockHash = hash[:]
	}

	// TODO(stgleb): Consider usage of O_DIRECT mode for writing
	file, err := os.OpenFile(config.BlockChain.DataFile, os.O_SYNC|os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)

	if err != nil {
		return nil, err
	}

	index, offset, err := NewIndex(config.BlockChain.DataFile)

	if err != nil {
		return nil, err
	}

	m := &BlockChain{
		file:          file,
		ticker:        time.NewTicker(time.Second * time.Duration(config.BlockChain.TimeOut)),
		DataFileName:  config.BlockChain.DataFile,
		Offset:        offset,
		Index:         index,
		LastBlockHash: prevBlockHash,
		BlockSize:     config.BlockChain.BlockSize,
		Timeout:       time.Duration(config.BlockChain.TimeOut) * time.Second,
		Input:         make(chan *Transaction),
		ShutDown:      make(chan chan struct{}),
		Search:        make(chan *SearchRequest),
	}

	go m.Run()

	return m, err
}

func (b BlockChain) Run() {
	var transactions = make([]Transaction, 0, b.BlockSize)

	for {
		select {
		case ch := <-b.ShutDown:
			GetLogger().Info("Shutdown blockchain")
			if err := b.Flush(transactions); err != nil {
				GetLogger().Error(err)
			}

			if err := b.file.Close(); err != nil {
				GetLogger().Error(err)
			}

			close(ch)
			return
		case tx := <-b.Input:
			GetLogger().Infof("Receive transaction %v", tx)
			transactions = append(transactions, *tx)

			if len(transactions) == b.BlockSize {
				if err := b.Flush(transactions); err != nil {
					GetLogger().Error(err)
				}
				// Reset ticket after transaction pool overflow
				b.ticker = time.NewTicker(b.Timeout)
				transactions = make([]Transaction, 0, b.BlockSize)
			}
		case <-b.ticker.C:
			GetLogger().Info("Flush by ticker")
			if err := b.Flush(transactions); err != nil {
				GetLogger().Error(err)
			}

			transactions = make([]Transaction, 0, b.BlockSize)
		case searchRequest := <-b.Search:
			GetLogger().Infof("Search by key %s", searchRequest.Key)

			go func() {
				var (
					err          error
					transactions []Transaction
				)

				// Search for key with in-memory inverted index and full scan of blockchain
				if b.IndexOn {
					transactions, err = b.Index.Get(searchRequest.Key)
				} else {
					transactions, err = FullScan(searchRequest.Key, b.DataFileName)
				}

				var errStr string

				if err != nil {
					errStr = err.Error()
				}
				searchResult := &SearchResult{
					transactions,
					errStr,
				}

				select {
				case <-searchRequest.ctx.Done():
					return
				case searchRequest.ResultChan <- searchResult:
				}
			}()
		}
	}
}

func (b BlockChain) Flush(transactions []Transaction) error {
	var block *Block

	// Do not create block and flush it on disk if there are no transactions
	if len(transactions) == 0 {
		GetLogger().Warn("Skip flushing empty block")
		return nil
	}

	block = NewBlock(b.LastBlockHash, transactions)
	blockBytes, err := json.Marshal(block)

	if err != nil {
		return err
	}

	blockSize := uint32(len(blockBytes))
	header := make([]byte, HEADER_SIZE)
	binary.LittleEndian.PutUint32(header, blockSize)

	// Append current block hash to the end of record to find it out after restart
	data := bytes.Join([][]byte{header, blockBytes, block.BlockHash}, []byte{})
	n, err := b.file.Write(data)

	if err != nil {
		return err
	}

	GetLogger().Debugf("Bytes written %d", n)
	err = b.file.Sync()

	if err != nil {
		return err
	}

	// Update index with block that was written to disk
	b.Index.Update(b.Offset, block)
	b.Offset += int64(len(data))

	b.LastBlockHash = block.BlockHash
	return nil
}
