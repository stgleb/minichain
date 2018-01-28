package minichain

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"testing"
)

// Test index get with padding
func TestInvertedIndexGet(t *testing.T) {
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

	padding := make([]byte, 256)
	data := bytes.Join([][]byte{padding, header, blockBytes, block.BlockHash}, []byte{})
	file := bytes.NewReader(data)

	index := &BloomFilterIndex{
		file: file,
		data: map[string][]int64{
			key: {
				256,
			},
		},
	}

	txs, err := index.Get(key)

	if err != nil {
		t.Error(err)
	}

	if len(txs) != 1 {
		t.Errorf("Wrong tranasction count expected %d actual %d", 1, len(txs))
	}
}

func TestInvertedIndexUpdate(t *testing.T) {
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
		data: map[string][]int64{},
	}

	index.Update(200, &block)

	if len(index.data[key]) != 1 {
		t.Errorf("tx with key %s was not added to index", key)
	}
}
