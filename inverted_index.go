package minichain

import (
	"io"
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

type InvertedIndex struct {
	// Mutex protects data map from races
	//TODO(stgleb): Consider using sync.Map
	m    sync.RWMutex
	data map[string][]int64
	file io.ReadSeeker
}

func NewInvertedIndex(file io.ReadSeeker) (Index, int64, error) {
	GetLogger().Info("Start building index")

	index := &InvertedIndex{
		file: file,
		data: make(map[string][]int64),
	}

	var (
		err        error
		blockCount int64
		offset     int64
		block      *Block
	)

	for {
		if block, offset, err = readBlock(file); err != nil {
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

func (index *InvertedIndex) Get(key string) ([]Transaction, error) {
	// Protect Get method with Write lock since it modifies Seeker e.g fd state.
	index.m.Lock()
	defer index.m.Unlock()

	offsets, ok := index.data[key]

	if !ok {
		return nil, KeyNotFoundErr
	}

	var (
		err          error
		transactions = make([]Transaction, 0, len(offsets))
	)

	for _, offset := range offsets {
		_, err = index.file.Seek(offset, 0)

		if err != nil {
			return nil, err
		}

		block, _, err := readBlock(index.file)

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
