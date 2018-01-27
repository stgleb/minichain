package minichain

import (
	"errors"
	"io"
	"os"
	"sync"
)

/*
	Inverted index that tracks offsets of blocks this key belongs to

	Example:

	blockchain

	0	block0: tx{key: hello}, tx{key: apple}
	28	block1: tx{key: banana}
	76	block2: tx{key: pear}, tx{key: hello}, tx{key: world}

	Index

		apple  - [0]
		banana - [28]
		hello  - [0, 76]
		pear   - [76]
		world  - [76]
*/

// TODO(stgleb): Use bloom filter instead of storing all keys in map.
type InvertedIndex struct {
	blockCount int
	fileName   string

	// Mutex protects data map from races
	//TODO(stgleb): Consider using sync.Map
	m    sync.RWMutex
	data map[string][]int64
}

var (
	KeyNotFoundErr   = errors.New("key not found")
	NotEnoughDataErr = errors.New("not enough data in file")
)

func NewIndex(fileName string) (*InvertedIndex, int64, error) {
	GetLogger().Infof("Start building index of %s", fileName)
	f, err := os.OpenFile(fileName, os.O_RDONLY, 0600)
	defer f.Close()

	index := &InvertedIndex{
		fileName: fileName,
		data:     make(map[string][]int64),
	}

	if err != nil {
		return index, 0, err
	}

	stats, err := f.Stat()

	if err != nil {
		return index, 0, err
	}

	if stats.Size() == 0 {
		return index, 0, nil
	}

	var (
		blockCount int64
		offset     int64
		block      *Block
	)

	for {
		if block, offset, err = readBlock(f); err != nil {
			if err != io.EOF {
				return nil, 0, err
			} else {
				break
			}
		} else {
			GetLogger().Debugf("Read block id %s on offset %d",
				string(block.BlockHash), offset)
			for _, tx := range block.Transactions {
				if index.data[string(tx.Key)] == nil {
					index.data[string(tx.Key)] = []int64{offset}
				} else {
					index.data[string(tx.Key)] = append(index.data[string(tx.Key)], offset)
				}
			}
			blockCount++
		}
	}

	GetLogger().Debugf("InvertedIndex has been built from %d blocks", blockCount)
	return index, offset, nil
}

// TODO(stgleb): Consider returning slice of tranasctions where such key was encountered
func (index *InvertedIndex) Get(key string) ([]Transaction, error) {
	f, err := os.OpenFile(index.fileName, os.O_RDONLY, 0600)
	defer f.Close()

	index.m.RLock()
	offsets, ok := index.data[key]
	// Make a copy of slice to avoid concurrent read-update
	offsets = offsets[:]
	index.m.RUnlock()

	if !ok {
		return nil, KeyNotFoundErr
	}

	var transactions = make([]Transaction, 0, len(offsets))

	for _, offset := range offsets {
		_, err = f.Seek(offset, 0)

		if err != nil {
			return nil, err
		}

		block, _, err := readBlock(f)

		if err != nil {
			return nil, err
		}

		for _, tx := range block.Transactions {
			if string(tx.Key) == key {
				transactions = append(transactions, tx)
			}
		}
	}

	return transactions, nil
}

// Update index with new transactions
func (index *InvertedIndex) Update(offset int64, block *Block) {
	for _, tx := range block.Transactions {
		index.m.Lock()
		// This update on slice that stores key offsets is safe since we allow only
		// one goroutine to update it.
		if index.data[string(tx.Key)] == nil {
			index.data[string(tx.Key)] = []int64{offset}
		} else {
			index.data[string(tx.Key)] = append(index.data[string(tx.Key)], offset)
		}
		index.m.Unlock()
	}
}
