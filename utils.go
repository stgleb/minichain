package minichain

import (
	"errors"
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
