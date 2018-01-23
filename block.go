package minichain

type Block struct {
	PrevBlockHash string        `json:"prev-block-hash"`
	BlockHash     string        `json:"block-hash"`
	Transactions  []Transaction `json:"transactions"`
}
