package minichain

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"strconv"
	"time"
)

type Block struct {
	Timestamp     int64
	PrevBlockHash []byte        `json:"prev-block-hash"`
	BlockHash     []byte        `json:"block-hash"`
	Transactions  []Transaction `json:"transactions"`
}

func NewBlock(prevBlockHash []byte, transactions []Transaction) *Block {
	var txHashes [][]byte
	var txHash [32]byte

	block := &Block{
		Timestamp:     time.Now().Unix(),
		PrevBlockHash: prevBlockHash,
		Transactions:  transactions,
	}

	for _, tx := range transactions {
		txHashes = append(txHashes, tx.Id)
	}

	timestampBytes := []byte(strconv.FormatInt(block.Timestamp, 10))
	txHashes = append(txHashes, timestampBytes)

	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))
	block.BlockHash = txHash[:]

	return block
}

func (b *Block) UnmarshalJSON(data []byte) error {
	type BlockAlias Block
	aux := &struct {
		*BlockAlias
	}{
		(*BlockAlias)(b),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		GetLogger().Error("Error while unmarshalling block")
		return err
	}

	return nil
}
