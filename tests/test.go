package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/valyala/fasthttp"
	"github.com/xuri/excelize/v2"
)

// BlockResponse represents the complete RPC response
type BlockResponse struct {
	Jsonrpc string       `json:"jsonrpc"`
	Id      int          `json:"id"`
	Result  BlockDetails `json:"result"`
}

// BlockDetails contains the main block information
type BlockDetails struct {
	BlockHeight       uint64        `json:"blockHeight"`
	BlockTime         *uint64       `json:"blockTime"` // Pointer because it can be null
	Blockhash         string        `json:"blockhash"`
	ParentSlot        uint64        `json:"parentSlot"`
	PreviousBlockhash string        `json:"previousBlockhash"`
	Transactions      []Transaction `json:"transactions"`
}

// Transaction represents a single transaction in the block
type Transaction struct {
	Meta        TransactionMeta    `json:"meta"`
	Transaction TransactionDetails `json:"transaction"`
	BlockTime   *uint64            `json:"blockTime,omitempty"`
	Slot        uint64             `json:"slot,omitempty"`
}

// TransactionMeta contains transaction metadata
type TransactionMeta struct {
	Err               interface{}        `json:"err"`
	Fee               uint64             `json:"fee"`
	PreBalances       []uint64           `json:"preBalances"`
	PostBalances      []uint64           `json:"postBalances"`
	InnerInstructions []InnerInstruction `json:"innerInstructions"`
	LogMessages       []string           `json:"logMessages"`
}

// InnerInstruction represents inner instructions within a transaction
type InnerInstruction struct {
	Index        uint32               `json:"index"`
	Instructions []InstructionDetails `json:"instructions"`
}

// InstructionDetails contains the details of an instruction
type InstructionDetails struct {
	ProgramId string   `json:"programId"`
	Program   string   `json:"program"`
	Data      string   `json:"data,omitempty"`
	Stack     []string `json:"stack,omitempty"`
}

// TransactionDetails contains the main transaction information
type TransactionDetails struct {
	Message    MessageDetails `json:"message"`
	Signatures []string       `json:"signatures"`
}

// MessageDetails contains the message part of a transaction
type MessageDetails struct {
	AccountKeys  []string             `json:"accountKeys"`
	Instructions []InstructionDetails `json:"instructions"`
}

// getBlockDetails fetches block information from the Solana network
func getBlockDetails(slot uint64, rpcURL string) (*BlockResponse, error) {
	reqBody := fmt.Sprintf(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "getBlock",
		"params": [
			%d,
			{
				"encoding": "json",
				"transactionDetails": "full",
				"rewards": false,
				"maxSupportedTransactionVersion": 0
			}
		]
	}`, slot)

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(rpcURL)
	req.Header.SetMethod("POST")
	req.Header.SetContentType("application/json")
	req.SetBodyString(reqBody)

	if err := fasthttp.Do(req, resp); err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}

	var blockResponse BlockResponse
	if err := json.Unmarshal(resp.Body(), &blockResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return &blockResponse, nil
}

// createExcelFile creates an Excel file with block details
func createExcelFile(block *BlockResponse) error {
	f := excelize.NewFile()

	// Create a new sheet
	sheetName := "BlockDetails"
	_, err := f.NewSheet(sheetName)
	if err != nil {
		return fmt.Errorf("failed to create sheet: %v", err)
	}

	// Write block information
	blockInfo := []struct {
		label string
		value interface{}
	}{
		{"Block Height", block.Result.BlockHeight},
		{"Block Hash", block.Result.Blockhash},
		{"Block Time", block.Result.BlockTime},
		{"Parent Slot", block.Result.ParentSlot},
		{"Previous Block Hash", block.Result.PreviousBlockhash},
		{"Number of Transactions", len(block.Result.Transactions)},
	}

	for i, info := range blockInfo {
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", i+1), info.label)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", i+1), info.value)
	}

	// Headers for transaction details starting from row 8
	headers := []string{
		"Transaction Index",
		"Signature",
		"Status",
		"Fee (lamports)",
		"Program IDs",
		"Account Keys",
		"Instruction Data",
		"Inner Instructions",
		"Balance Changes",
	}

	// Set headers
	for i, header := range headers {
		col := string(rune('A' + i))
		f.SetCellValue(sheetName, fmt.Sprintf("%s8", col), header)
	}

	// Write transaction details
	rowIndex := 9
	for i, tx := range block.Result.Transactions {
		// Basic transaction info
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowIndex), i+1)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowIndex), tx.Transaction.Signatures[0])

		status := "Success"
		if tx.Meta.Err != nil {
			status = "Failed"
		}
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowIndex), status)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowIndex), tx.Meta.Fee)

		// Program IDs from instructions
		var programIDs []string
		for _, inst := range tx.Transaction.Message.Instructions {
			programIDs = append(programIDs, inst.ProgramId)
		}
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowIndex), strings.Join(programIDs, "\n"))

		// Account keys
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowIndex), strings.Join(tx.Transaction.Message.AccountKeys, "\n"))

		// Instruction data
		var instData []string
		for j, inst := range tx.Transaction.Message.Instructions {
			instData = append(instData, fmt.Sprintf("Inst %d: %s", j, inst.Data))
		}
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", rowIndex), strings.Join(instData, "\n"))

		// Inner instructions
		var innerInsts []string
		for _, inner := range tx.Meta.InnerInstructions {
			for j, inst := range inner.Instructions {
				innerInsts = append(innerInsts, fmt.Sprintf("Inner %d: Program=%s, Data=%s", j, inst.ProgramId, inst.Data))
			}
		}
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", rowIndex), strings.Join(innerInsts, "\n"))

		// Balance changes
		var balanceChanges []string
		for j, key := range tx.Transaction.Message.AccountKeys {
			if j < len(tx.Meta.PreBalances) && j < len(tx.Meta.PostBalances) {
				preBalance := tx.Meta.PreBalances[j]
				postBalance := tx.Meta.PostBalances[j]
				if preBalance != postBalance {
					change := postBalance - preBalance
					balanceChanges = append(balanceChanges, fmt.Sprintf("%s: %d → %d (Δ%d)",
						key, preBalance, postBalance, change))
				}
			}
		}
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", rowIndex), strings.Join(balanceChanges, "\n"))

		rowIndex++
	}

	// Adjust column width for better readability
	for i := 0; i < len(headers); i++ {
		col := string(rune('A' + i))
		f.SetColWidth(sheetName, col, col, 40)
	}

	// Save file
	if err := f.SaveAs("block_details.xlsx"); err != nil {
		return fmt.Errorf("failed to save excel file: %v", err)
	}

	return nil
}

func main() {
	// Load environment variables
	if err := godotenv.Load("../.env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	rpcURL := os.Getenv("RPC_URL")
	if rpcURL == "" {
		log.Fatal("RPC_URL not set in .env file")
	}

	// Get block slot from user input
	var slot uint64
	fmt.Print("Enter Solana block slot number: ")
	if _, err := fmt.Scan(&slot); err != nil {
		log.Fatal("Invalid slot number:", err)
	}

	// Fetch block details
	blockResponse, err := getBlockDetails(slot, rpcURL)
	if err != nil {
		log.Fatal("Failed to get block details:", err)
	}

	// Create Excel file with the block details
	if err := createExcelFile(blockResponse); err != nil {
		log.Fatal("Failed to create Excel file:", err)
	}
	fmt.Println("\nBlock details have been saved to block_details.xlsx")

	// Print basic block information
	fmt.Printf("\nBlock Information:\n")
	fmt.Printf("Block Height: %d\n", blockResponse.Result.BlockHeight)
	if blockResponse.Result.BlockTime != nil {
		fmt.Printf("Block Time: %d\n", *blockResponse.Result.BlockTime)
	}
	fmt.Printf("Block Hash: %s\n", blockResponse.Result.Blockhash)
	fmt.Printf("Parent Slot: %d\n", blockResponse.Result.ParentSlot)
	fmt.Printf("Previous Block Hash: %s\n", blockResponse.Result.PreviousBlockhash)
	fmt.Printf("Number of Transactions: %d\n", len(blockResponse.Result.Transactions))
}
