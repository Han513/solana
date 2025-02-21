package models

type BlockResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	Result  struct {
		BlockHeight       uint64        `json:"blockHeight"`
		BlockTime         *uint64       `json:"blockTime"`
		Blockhash         string        `json:"blockhash"`
		ParentSlot        uint64        `json:"parentSlot"`
		PreviousBlockhash string        `json:"previousBlockhash"`
		Transactions      []Transaction `json:"transactions"`
	} `json:"result"`
	Id int `json:"id"`
}

type BlockMessage struct {
	BlockHeight       uint64            `json:"blockHeight"`
	BlockTime         *uint64           `json:"blockTime"`
	Blockhash         string            `json:"blockhash"`
	ParentSlot        uint64            `json:"parentSlot"`
	PreviousBlockhash string            `json:"previousBlockhash"`
	Transactions      []TransactionInfo `json:"transactions"`
	Timestamp         int64             `json:"timestamp"`
}
