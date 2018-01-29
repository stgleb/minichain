package minichain

import (
	"testing"
	"crypto/sha256"
	"encoding/json"
	"encoding/binary"
	"bytes"
	"github.com/willf/bloom"
	"math"
)

// Test index get with padding
func TestBloomFilterIndexGet(t *testing.T) {
	genesis := sha256.Sum256([]byte(GENESIS_BLOCK))
	hash := sha256.Sum256([]byte("hash seed"))
	key := "key-1"

	block := Block{
		Timestamp:     1,
		PrevBlockHash: genesis[:],
		BlockHash:     hash[:],
		Transactions: []Transaction{
			{
				Id:        []byte("dummy-txs-2"),
				Timestamp: 1,
				Key:       key,
				Value:     "value1",
			},
			{
				Id:        []byte("dummy-txs-2"),
				Timestamp: 1,
				Key:       "key2",
				Value:     "value2",
			},
		},
	}

	blockBytes, err := json.Marshal(block)

	if err != nil {
		t.Error(err)
	}

	blockSize := uint32(len(blockBytes))
	header := make([]byte, HEADER_SIZE)
	binary.LittleEndian.PutUint32(header, blockSize)

	offset := int64(256)
	padding := make([]byte, offset)
	data := bytes.Join([][]byte{padding, header, blockBytes, block.BlockHash}, []byte{})
	file := bytes.NewReader(data)
	filter := bloom.New(uint(len(block.Transactions)), uint(math.Ceil(0.69*20)))
	filter.AddString(key)

	info := &BlockInfo{
		offset: offset,
		filter: filter,
	}

	index := &BloomFilterIndex{
		blocks: []*BlockInfo{info},
		file:file,
	}

	txs, err := index.Get(key)

	if err != nil {
		t.Error(err)
	}

	if len(txs) != 1 {
		t.Errorf("Wrong tranasction count expected %d actual %d", 1, len(txs))
	}
}

func TestBloomFilterIndexUpdate(t *testing.T) {
	genesis := sha256.Sum256([]byte(GENESIS_BLOCK))
	hash := sha256.Sum256([]byte("hash seed"))
	key := "key1"

	block := Block{
		Timestamp:     1,
		PrevBlockHash: genesis[:],
		BlockHash:     hash[:],
		Transactions: []Transaction{
			{
				Id:        []byte("dummy-txs-2"),
				Timestamp: 1,
				Key:       key,
				Value:     "value1",
			},
			{
				Id:        []byte("dummy-txs-2"),
				Timestamp: 1,
				Key:       "key2",
				Value:     "value2",
			},
		},
	}

	index := &BloomFilterIndex{
		blocks: make([]*BlockInfo, 0),
	}

	index.Update(200, &block)

	if len(index.blocks) != 1 {
		t.Errorf("Index was not updated")
	}

	if !index.blocks[0].filter.TestString(key) {
		t.Errorf("Cannot find key in index")
	}
}

