package minichain

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"os"
)

type Index struct {
	blockCount int
	file       *os.File
	data       map[string]int64
}

var (
	KeyNotFoundErr   = errors.New("key not found")
	NotEnoughDataErr = errors.New("not enough data in file")
)

func NewIndex(fileName string) (*Index, int64, error) {
	GetLogger().Infof("Start building index of %s", fileName)
	f, err := os.OpenFile(fileName, os.O_RDONLY, 0600)

	index := &Index{
		file: f,
		data: make(map[string]int64),
	}

	if err != nil {
		return index, -1, err
	}

	stats, err := f.Stat()

	if err != nil {
		return index, -1, err
	}

	if stats.Size() == 0 {
		return index, -1, nil
	}

	var offset int64
	var block *Block

	for {
		if block, offset, err = readBlock(index.file); err != nil {
			if err != io.EOF {
				return nil, -1, err
			} else {
				break
			}
		} else {
			for _, tx := range block.Transactions {
				index.data[string(tx.Key)] = offset
			}
		}
	}

	return index, offset, nil
}

func (index *Index) Get(key string) (*Transaction, error) {
	offset, ok := index.data[key]

	if !ok {
		return nil, KeyNotFoundErr
	}

	_, err := index.file.Seek(offset, 0)

	if err != nil {
		return nil, err
	}

	block, _, err := readBlock(index.file)

	if err != nil {
		return nil, err
	}

	for _, tx := range block.Transactions {
		if string(tx.Key) == key {
			return tx, nil
		}
	}

	return nil, KeyNotFoundErr
}

// Update index with new transactions
func (index *Index) Update(offset int64, block *Block) {
	for _, tx := range block.Transactions {
		index.data[string(tx.Key)] = offset
	}
}

// Reads block from blockchain file, assumes that file pointer of fd is set on
// the beggining of next block
func readBlock(fd *os.File) (*Block, int64, error) {
	offset, err := fd.Seek(0, 1)

	if err != nil {
		return nil, -1, err
	}

	headerData := make([]byte, 0, 4)
	n, err := fd.Read(headerData)

	if err != nil {
		return nil, -1, err
	}

	if n != HEADER_SIZE {
		return nil, -1, NotEnoughDataErr
	}

	blockSize := binary.LittleEndian.Uint32(headerData)
	blockBuffer := make([]byte, 0, blockSize)
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

	var block *Block

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
