package minichain

import (
	"crypto/sha256"
	"encoding/json"
	"time"
)

type Transaction struct {
	Id        string `json:"id"`
	Key       string `json:"key"`
	Value     string `json:"value"`
	Timestamp int64  `json:"timestamp"`
}

func NewTransaction(key, value string) (*Transaction, error) {
	tx := &Transaction{
		Key:       key,
		Value:     value,
		Timestamp: time.Now().Unix(),
	}

	if data, err := json.Marshal(tx); err != nil {
		hash := sha256.Sum256(data)
		tx.Id = string(hash[:])

		return tx, nil
	} else {
		return nil, err
	}
}
