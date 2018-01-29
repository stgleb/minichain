package minichain

import (
	"errors"
	"io"
)

const (
	INVERTED_INDEX = "InvertedIndex"
	BLOOM_FILTER   = "BloomFilter"
)

type Index interface {
	Get(string) ([]Transaction, error)
	Update(int64, *Block)
}

var (
	KeyNotFoundErr   = errors.New("key not found")
	NotEnoughDataErr = errors.New("not enough data in reader")
)

func NewIndex(reader io.ReadSeeker, indexType string) (Index, int64, error) {
	switch indexType {
	case INVERTED_INDEX:
		return NewInvertedIndex(reader)
	case BLOOM_FILTER:
		return NewBloomFilterIndex(reader)
	default:
		return nil, 0, nil
	}
}
