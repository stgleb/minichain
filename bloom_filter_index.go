package minichain

import (
	"github.com/willf/bloom"
	"io"
	"math"
	"sync"
)

/*
	Bloom filter index stores small set of information about the block and
	saves disk read operations.
*/

type BlockInfo struct {
	filter *bloom.BloomFilter
	offset int64
}

type BloomFilterIndex struct {
	// Mutex protects slice and seeker from updates
	m      sync.RWMutex
	blocks []*BlockInfo
	file   io.ReadSeeker
}

func NewBloomFilterIndex(file io.ReadSeeker) (Index, int64, error) {
	GetLogger().Info("Start building index")

	index := &BloomFilterIndex{
		file:   file,
		blocks: make([]*BlockInfo, 0, 32),
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
			// Such parameter of hash function count minimizes false positive rate minimizes
			filter := bloom.New(uint(len(block.Transactions)), uint(math.Ceil(0.69*20)))
			info := &BlockInfo{
				filter: filter,
				offset: offset,
			}

			for _, tx := range block.Transactions {
				info.filter.AddString(tx.Key)
			}
			// Append info about the block
			index.blocks = append(index.blocks, info)
			blockCount++
		}
	}

	GetLogger().Debugf("BloomFilterIndex has been built from %d blocks", blockCount)
	return index, offset, nil
}

func (index *BloomFilterIndex) Get(key string) ([]Transaction, error) {
	// Protect Get method with Write lock since it modifies Seeker e.g fd state.
	index.m.Lock()
	defer index.m.Unlock()

	var (
		err          error
		transactions = make([]Transaction, 0)
	)

	for _, blockInfo := range index.blocks {
		// If block doesn't contain the key
		if !blockInfo.filter.TestString(key) {
			continue
		}

		offset := blockInfo.offset
		_, err = index.file.Seek(offset, 0)

		if err != nil {
			return nil, err
		}

		block, _, err := readBlock(index.file)

		if err != nil {
			return nil, err
		}

		// Check whether block contains key or not
		for _, tx := range block.Transactions {
			if string(tx.Key) == key {
				transactions = append(transactions, tx)
			}
		}
	}

	return transactions, nil
}

// Update index with new transactions
func (index *BloomFilterIndex) Update(offset int64, block *Block) {
	// Such parameter of hash functions gives least false positive probability 0.0001
	filter := bloom.New(uint(len(block.Transactions)) * 20, uint(math.Ceil(0.69*20)))
	info := &BlockInfo{
		filter: filter,
		offset: offset,
	}

	for _, tx := range block.Transactions {
		info.filter.AddString(tx.Key)
	}

	// Protect slice with lock for update
	index.m.Lock()
	index.blocks = append(index.blocks, info)
	index.m.Unlock()

}
