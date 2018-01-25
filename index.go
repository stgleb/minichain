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
	fileName   string
	data       map[string]int64
}

var (
	KeyNotFoundErr   = errors.New("key not found")
	NotEnoughDataErr = errors.New("not enough data in file")
)

func NewIndex(fileName string) (*Index, int64, error) {
	GetLogger().Infof("Start building index of %s", fileName)
	f, err := os.OpenFile(fileName, os.O_RDONLY, 0600)
	defer f.Close()

	index := &Index{
		fileName: fileName,
		data:     make(map[string]int64),
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
				index.data[string(tx.Key)] = offset
			}
			blockCount++
		}
	}

	GetLogger().Debugf("Index has been built from %d blocks", blockCount)
	return index, offset, nil
}

func (index *Index) Get(key string) (*Transaction, error) {
	f, err := os.OpenFile(index.fileName, os.O_RDONLY, 0600)
	defer f.Close()
	offset, ok := index.data[key]

	if !ok {
		return nil, KeyNotFoundErr
	}

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
			return &tx, nil
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
