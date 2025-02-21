package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/IBM/sarama"
)

type BlockMessage struct {
	BlockHeight       uint64            `json:"blockHeight"`
	BlockTime         *uint64           `json:"blockTime"`
	Blockhash         string            `json:"blockhash"`
	ParentSlot        uint64            `json:"parentSlot"`
	PreviousBlockhash string            `json:"previousBlockhash"`
	Transactions      []TransactionInfo `json:"transactions"`
	Timestamp         int64             `json:"timestamp"`
}

type TransactionInfo struct {
	Signature      string          `json:"signature"`
	Status         string          `json:"status"`
	Fee            uint64          `json:"fee"`
	AccountKeys    []string        `json:"accountKeys"`
	Instructions   []Instruction   `json:"instructions"`
	BalanceChanges []BalanceChange `json:"balanceChanges"`
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

func prettyPrintBlockMessage(msg BlockMessage) {
	fmt.Println("\n=== Block Information ===")
	fmt.Printf("Block Height: %d\n", msg.BlockHeight)
	if msg.BlockTime != nil {
		fmt.Printf("Block Time: %v\n", time.Unix(int64(*msg.BlockTime), 0))
	}
	fmt.Printf("Block Hash: %s\n", msg.Blockhash)
	fmt.Printf("Parent Slot: %d\n", msg.ParentSlot)
	fmt.Printf("Previous Block Hash: %s\n", msg.PreviousBlockhash)
	fmt.Printf("Message Timestamp: %v\n", time.Unix(msg.Timestamp, 0))
	fmt.Printf("Transaction Count: %d\n", len(msg.Transactions))

	fmt.Println("\n=== Transactions ===")
	for i, tx := range msg.Transactions {
		fmt.Printf("\nTransaction %d:\n", i+1)
		fmt.Printf("  Signature: %s\n", tx.Signature)
		fmt.Printf("  Status: %s\n", tx.Status)
		fmt.Printf("  Fee: %d lamports (%.9f SOL)\n", tx.Fee, float64(tx.Fee)/1e9)

		if len(tx.Instructions) > 0 {
			fmt.Printf("\n  Instructions:\n")
			for j, inst := range tx.Instructions {
				fmt.Printf("    [%d] Program: %s\n", j+1, inst.ProgramId)
				fmt.Printf("        Data: %s\n", inst.Data)
				fmt.Printf("        Accounts (%d):\n", len(inst.Accounts))
				for k, acc := range inst.Accounts {
					fmt.Printf("          %d. %s\n", k+1, acc)
				}
			}
		}

		if len(tx.BalanceChanges) > 0 {
			fmt.Printf("\n  Balance Changes:\n")
			for _, change := range tx.BalanceChanges {
				fmt.Printf("    Account: %s\n", change.Account)
				fmt.Printf("      Change: %d lamports (%.9f SOL)\n", change.Change, float64(change.Change)/1e9)
				fmt.Printf("      Pre  : %.9f SOL\n", float64(change.PreBalance)/1e9)
				fmt.Printf("      Post : %.9f SOL\n", float64(change.PostBalance)/1e9)
			}
		}

		fmt.Println("  " + strings.Repeat("-", 50))
	}
	fmt.Println("\n=== End of Block ===")
}

func createKafkaClient() (sarama.Client, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_8_0_0
	config.Net.DialTimeout = time.Second * 30
	config.Net.ReadTimeout = time.Second * 30
	config.Net.WriteTimeout = time.Second * 30

	client, err := sarama.NewClient([]string{"127.0.0.1:8998"}, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	return client, nil
}

func createKafkaConsumer(client sarama.Client) (sarama.Consumer, error) {
	consumer, err := sarama.NewConsumerFromClient(client)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %v", err)
	}

	return consumer, nil
}

func getPartitionOffsets(client sarama.Client, topic string, partition int32) (oldest, newest int64, err error) {
	oldest, err = client.GetOffset(topic, partition, sarama.OffsetOldest)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get oldest offset: %v", err)
	}

	newest, err = client.GetOffset(topic, partition, sarama.OffsetNewest)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get newest offset: %v", err)
	}

	return oldest, newest, nil
}

func main() {
	// 創建 client
	client, err := createKafkaClient()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// 創建 consumer
	consumer, err := createKafkaConsumer(client)
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}
	defer consumer.Close()

	topic := "solana"
	partition := int32(0)

	// 獲取最舊和最新的 offset
	oldest, newest, err := getPartitionOffsets(client, topic, partition)
	if err != nil {
		log.Fatalf("Failed to get offsets: %v", err)
	}

	fmt.Printf("\nAvailable message range for topic '%s', partition %d:\n", topic, partition)
	fmt.Printf("Oldest offset: %d\n", oldest)
	fmt.Printf("Newest offset: %d\n", newest)
	fmt.Printf("Total messages: %d\n\n", newest-oldest)

	var choice int
	var startOffset int64
	fmt.Println("Choose how to read messages:")
	fmt.Println("1. Read from beginning")
	fmt.Println("2. Read only new messages")
	fmt.Println("3. Read from specific offset")
	fmt.Print("Enter your choice (1-3): ")
	fmt.Scan(&choice)

	switch choice {
	case 1:
		startOffset = sarama.OffsetOldest
	case 2:
		startOffset = sarama.OffsetNewest
	case 3:
		fmt.Printf("Enter offset (between %d and %d): ", oldest, newest)
		var userOffset int64
		fmt.Scan(&userOffset)
		if userOffset < oldest || userOffset > newest {
			log.Fatalf("Invalid offset. Must be between %d and %d", oldest, newest)
		}
		startOffset = userOffset
	default:
		log.Fatal("Invalid choice")
	}

	partitionConsumer, err := consumer.ConsumePartition(topic, partition, startOffset)
	if err != nil {
		log.Fatalf("Failed to create partition consumer: %v", err)
	}
	defer partitionConsumer.Close()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	fmt.Println("\nConsumer started. Waiting for messages...")
	fmt.Println("Press Ctrl+C to exit")
	fmt.Println("----------------------------------------")

	consumed := 0
ConsumerLoop:
	for {
		select {
		case msg := <-partitionConsumer.Messages():
			consumed++
			fmt.Printf("\nReceived message at offset %d:\n", msg.Offset)

			// 解析消息
			var blockMsg BlockMessage
			if err := json.Unmarshal(msg.Value, &blockMsg); err != nil {
				log.Printf("Error parsing message: %v\n", err)
				continue
			}

			// 格式化輸出
			prettyPrintBlockMessage(blockMsg)
			fmt.Println("----------------------------------------")

		case err := <-partitionConsumer.Errors():
			log.Printf("Error: %v\n", err)

		case <-signals:
			break ConsumerLoop
		}
	}

	fmt.Printf("\nConsumed %d messages\n", consumed)
}
