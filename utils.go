package minichain

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"os"
)

func GetLastBlockHash(fileName string) ([]byte, error) {
	f, err := os.OpenFile(fileName, os.O_RDONLY, 0600)

	if err != nil {
		return nil, err
	}

	_, err = f.Seek(-DIGEST_SIZE, 2)

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

func FullScan(key, fileName string) ([]Transaction, error) {
	f, err := os.OpenFile(fileName, os.O_RDONLY, 0600)

	if err != nil {
		return nil, err
	}

	var (
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

// Reads block from blockchain file, assumes that file pointer of fd is set on
// the beginning of next block
func readBlock(fd *os.File) (*Block, int64, error) {
	offset, err := fd.Seek(0, 1)

	if err != nil {
		return nil, -1, err
	}

	headerData := make([]byte, HEADER_SIZE)
	n, err := fd.Read(headerData)

	if err != nil {
		return nil, -1, err
	}

	if n != HEADER_SIZE {
		return nil, -1, NotEnoughDataErr
	}

	blockSize := binary.LittleEndian.Uint32(headerData)
	blockBuffer := make([]byte, blockSize)
	n, err = fd.Read(blockBuffer)

	if err != nil {
		return nil, -1, err
	}

	if uint32(n) != blockSize {
		return nil, -1, NotEnoughDataErr
	}

	if err != nil {
		return nil, -1, err
	}

	var block = &Block{}

	err = json.Unmarshal(blockBuffer, block)

	if err != nil {
		return nil, -1, err
	}

	// Set fd to begin of next block or EOF
	_, err = fd.Seek(DIGEST_SIZE, 1)

	if err != nil {
		return nil, -1, err
	}

	return block, offset, nil
}
