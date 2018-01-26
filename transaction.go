package minichain

import (
	"bytes"
	"crypto/sha256"
	"strconv"
	"time"
)

type Transaction struct {
	Id        []byte `json:"id"`
	Key       string `json:"key"`
	Value     string `json:"value"`
	Timestamp int64  `json:"timestamp"`
}

func NewTransaction(key, value string) *Transaction {
	tx := &Transaction{
		Key:       key,
		Value:     value,
		Timestamp: time.Now().Unix(),
	}

	timestampBytes := []byte(strconv.FormatInt(tx.Timestamp, 10))

	header := bytes.Join([][]byte{[]byte(key), []byte(value), timestampBytes}, []byte{})
	hash := sha256.Sum256(header)
	tx.Id = hash[:]

	return tx
}
