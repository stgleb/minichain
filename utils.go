package minichain

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"sync"
)

var (
	m sync.Mutex
)

func getLastBlockHash(f io.ReadSeeker) ([]byte, error) {
	_, err := f.Seek(-DIGEST_SIZE, 2)

	if err != nil {
		return nil, err
	}

	prevBlockHash := make([]byte, DIGEST_SIZE)
	n, err := f.Read(prevBlockHash)

	if err != nil {
		return nil, err
	}

	if n != DIGEST_SIZE {
		return nil, errors.New("not enough bytes to read")
	}

	return prevBlockHash, nil
}

func fullScan(key string, f io.ReadSeeker) ([]Transaction, error) {
	var (
		err          error
		blockCount   int64
		offset       int64
		block        *Block
		transactions = make([]Transaction, 0)
	)

	for {
		if block, offset, err = readBlock(f); err != nil {
			if err != io.EOF {
				return nil, err
			} else {
				break
			}
		} else {
			GetLogger().Debugf("Read block id %s on offset %d",
				string(block.BlockHash), offset)
			for _, tx := range block.Transactions {
				if tx.Key == key {
					transactions = append(transactions, tx)
				}
			}
			blockCount++
		}
	}

	if len(transactions) == 0 {
		return nil, KeyNotFoundErr
	}

	return transactions, nil

}

// Reads block from blockchain writer, assumes that writer pointer of fd is set on
// the beginning of next block
func readBlock(reader io.ReadSeeker) (*Block, int64, error) {
	// Protect function with lock since it modifies reader state
	m.Lock()
	defer m.Unlock()

	offset, err := reader.Seek(0, 1)

	if err != nil {
		return nil, offset, err
	}

	headerData := make([]byte, HEADER_SIZE)
	n, err := reader.Read(headerData)

	if err != nil {
		return nil, 0, err
	}

	if n != HEADER_SIZE {
		return nil, offset, NotEnoughDataErr
	}

	blockSize := binary.LittleEndian.Uint32(headerData)
	blockBuffer := make([]byte, blockSize)
	n, err = reader.Read(blockBuffer)

	if err != nil {
		return nil, offset, err
	}

	if uint32(n) != blockSize {
		return nil, offset, NotEnoughDataErr
	}

	if err != nil {
		return nil, offset, err
	}

	var block = &Block{}

	err = json.Unmarshal(blockBuffer, block)

	if err != nil {
		return nil, offset, err
	}

	// Set reader to begin of next block or EOF
	_, err = reader.Seek(DIGEST_SIZE, 1)

	if err != nil {
		return nil, offset, err
	}

	return block, offset, nil
}
