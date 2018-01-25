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

func NewIndex(fileName string) (*Index, error) {
	f, err := os.OpenFile(fileName, os.O_RDONLY, 0600)

	stats, err := f.Stat()

	if err != nil {
		return nil, err
	}

	index := &Index{
		file: f,
		data: make(map[string]int64),
	}

	if stats.Size() == 0 {
		return index, nil
	}

	for {
		if block, offset, err := index.readBlock(); err != nil {
			if err != io.EOF {
				return nil, err
			} else {
				break
			}
		} else {
			for _, tx := range block.Transactions {
				index.data[string(tx.Key)] = offset
			}
		}
	}

	return index, nil
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

	block, _, err := index.readBlock()

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

func (index *Index) Update(offset int64, block *Block) {
	for _, tx := range block.Transactions {
		index.data[string(tx.Key)] = offset
	}
}

func (index *Index) readBlock() (*Block, int64, error) {
	offset, err := index.file.Seek(0, 1)

	if err != nil {
		return nil, -1, err
	}

	headerData := make([]byte, 0, 4)
	n, err := index.file.Read(headerData)

	if err != nil {
		return nil, -1, err
	}

	if n != HEADER_SIZE {
		return nil, -1, NotEnoughDataErr
	}

	blockSize := binary.LittleEndian.Uint32(headerData)
	blockBuffer := make([]byte, 0, blockSize)
	n, err = index.file.Read(blockBuffer)

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

	_, err = index.file.Seek(DIGEST_SIZE, 1)

	if err != nil {
		return nil, -1, err
	}

	return block, offset, nil
}
