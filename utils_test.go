package minichain

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"testing"
)

func TestGetLastBlockHash(t *testing.T) {
	backedSlice := sha256.Sum256([]byte(GENESIS_BLOCK))
	buffer := bytes.NewReader(backedSlice[:])

	hash, err := getLastBlockHash(buffer)

	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(hash, backedSlice[:]) {
		t.Errorf("Hash from file %v is not equal to actual hash %v", hash, backedSlice)
	}
}

func TestFullScan(t *testing.T) {
	key := "key"
	expectedTxs := 2

	genesis := sha256.Sum256([]byte(GENESIS_BLOCK))
	block1Hash := sha256.Sum256([]byte("hash0"))
	block2Hash := sha256.Sum256([]byte("hash1"))
	block3Hash := sha256.Sum256([]byte("hash2"))

	blockChain := []Block{
		{
			Timestamp:     1,
			PrevBlockHash: genesis[:],
			BlockHash:     block1Hash[:],
			Transactions: []Transaction{
				{
					Id:        []byte("dummy-tx"),
					Timestamp: 1,
					Key:       "key1",
					Value:     "value1",
				},
			},
		},
		{
			Timestamp:     2,
			PrevBlockHash: block1Hash[:],
			BlockHash:     block2Hash[:],
			Transactions: []Transaction{
				{
					Id:        []byte("dummy-tx"),
					Timestamp: 2,
					Key:       key,
					Value:     "value2",
				},
			},
		},
		{
			Timestamp:     3,
			PrevBlockHash: block2Hash[:],
			BlockHash:     block3Hash[:],
			Transactions: []Transaction{
				{
					Id:        []byte("dummy-tx"),
					Timestamp: 3,
					Key:       "key3",
					Value:     "value3",
				},
				{
					Id:        []byte("dummy-tx"),
					Timestamp: 4,
					Key:       key,
					Value:     "value4",
				},
			},
		},
	}

	backedArray := make([]byte, 0)

	for _, block := range blockChain {
		blockBytes, err := json.Marshal(block)

		if err != nil {
			t.Error(err)
		}

		blockSize := uint32(len(blockBytes))
		header := make([]byte, HEADER_SIZE)
		binary.LittleEndian.PutUint32(header, blockSize)

		data := bytes.Join([][]byte{header, blockBytes, block.BlockHash}, []byte{})
		backedArray = append(backedArray, data...)
	}

	file := bytes.NewReader(backedArray)
	txs, err := fullScan(key, file)

	if err != nil {
		t.Error(err)
	}

	if len(txs) != expectedTxs {
		t.Errorf("Expected to find %d transactionas actual %d", expectedTxs, len(txs))
	}

}

func TestReadBlock(t *testing.T) {
	genesis := sha256.Sum256([]byte(GENESIS_BLOCK))
	hash := sha256.Sum256([]byte("hash seed"))

	block := Block{
		Timestamp:     1,
		PrevBlockHash: genesis[:],
		BlockHash:     hash[:],
		Transactions: []Transaction{
			{
				Id:        []byte("dummy-tx-2"),
				Timestamp: 1,
				Key:       "key1",
				Value:     "value1",
			},
			{
				Id:        []byte("dummy-tx-2"),
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

	data := bytes.Join([][]byte{header, blockBytes, block.BlockHash}, []byte{})
	reader := bytes.NewReader(data)

	b, offset, err := readBlock(reader)

	if err != nil {
		t.Error(err)
	}

	if offset != 0 {
		t.Errorf("Expected offset %d actual %d",
			0, offset)
	}

	if len(b.Transactions) != len(b.Transactions) {
		t.Errorf("Expected len of transactions %d actual %d",
			len(block.Transactions), len(b.Transactions))
	}
}
