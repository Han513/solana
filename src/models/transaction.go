package models

type Transaction struct {
	Meta struct {
		Err               interface{}    `json:"err"`
		Fee               uint64         `json:"fee"`
		PreBalances       []uint64       `json:"preBalances"`
		PostBalances      []uint64       `json:"postBalances"`
		PreTokenBalances  []TokenBalance `json:"preTokenBalances"`
		PostTokenBalances []TokenBalance `json:"postTokenBalances"`
		InnerInstructions []struct {
			Index        uint32 `json:"index"`
			Instructions []struct {
				Accounts    []uint64 `json:"accounts"`
				Data        string   `json:"data"`
				ProgramId   string   `json:"programId"`
				StackHeight *uint32  `json:"stackHeight,omitempty"`
			} `json:"instructions"`
		} `json:"innerInstructions"`
		LogMessages []string `json:"logMessages"`
		Status      struct {
			Ok interface{} `json:"Ok"`
		} `json:"status"`
		ComputeUnitsConsumed uint64 `json:"computeUnitsConsumed"`
		LoadedAddresses      struct {
			Readonly []string `json:"readonly"`
			Writable []string `json:"writable"`
		} `json:"loadedAddresses"`
	} `json:"meta"`
	Transaction struct {
		Message struct {
			AccountKeys []string `json:"accountKeys"`
			Header      struct {
				NumRequiredSignatures       uint8 `json:"numRequiredSignatures"`
				NumReadonlySignedAccounts   uint8 `json:"numReadonlySignedAccounts"`
				NumReadonlyUnsignedAccounts uint8 `json:"numReadonlyUnsignedAccounts"`
			} `json:"header"`
			Instructions []struct {
				Accounts       []uint64 `json:"accounts"`
				Data           string   `json:"data"`
				ProgramIdIndex uint8    `json:"programIdIndex"`
				StackHeight    *uint32  `json:"stackHeight,omitempty"`
			} `json:"instructions"`
			RecentBlockhash     string `json:"recentBlockhash"`
			AddressTableLookups []struct {
				AccountKey      string  `json:"accountKey"`
				WritableIndexes []uint8 `json:"writableIndexes"`
				ReadonlyIndexes []uint8 `json:"readonlyIndexes"`
			} `json:"addressTableLookups,omitempty"`
		} `json:"message"`
		Signatures []string `json:"signatures"`
	} `json:"transaction"`
	Version interface{} `json:"version,omitempty"`
}

type TransactionInfo struct {
	Signature      string          `json:"signature"`
	Status         string          `json:"status"`
	Fee            uint64          `json:"fee"`
	AccountKeys    []string        `json:"accountKeys"`
	Instructions   []Instruction   `json:"instructions"`
	BalanceChanges []BalanceChange `json:"balanceChanges"`
	TokenBalances  []TokenBalance  `json:"tokenBalances"`
	ComputeUnits   uint64          `json:"computeUnits"`
	LogMessages    []string        `json:"logMessages"`
}

type Instruction struct {
	ProgramId string   `json:"programId"`
	Data      string   `json:"data"`
	Accounts  []string `json:"accounts"`
}

type BalanceChange struct {
	Account     string `json:"account"`
	PreBalance  uint64 `json:"preBalance"`
	PostBalance uint64 `json:"postBalance"`
	Change      int64  `json:"change"`
}

type TokenBalance struct {
	AccountIndex  uint64 `json:"accountIndex"`
	Mint          string `json:"mint"`
	Owner         string `json:"owner"`
	ProgramId     string `json:"programId"`
	UITokenAmount struct {
		Amount         string  `json:"amount"`
		Decimals       uint8   `json:"decimals"`
		UIAmount       float64 `json:"uiAmount"`
		UIAmountString string  `json:"uiAmountString"`
	} `json:"uiTokenAmount"`
}
