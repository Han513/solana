package services

import (
	"encoding/json"
	"fmt"
	"time"

	"solana/src/config"
	"solana/src/models"

	"github.com/IBM/sarama"
)

type KafkaProducer struct {
	producer sarama.SyncProducer
	Topic    string
}

func NewKafkaProducer(config *config.KafkaConfig, brokers []string) (*KafkaProducer, error) {
	producer, err := sarama.NewSyncProducer(brokers, config.ToSaramaConfig())
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %v", err)
	}

	return &KafkaProducer{
		producer: producer,
		// 可以將 topic 設為 KafkaProducer 結構的屬性，並在這裡初始化
	}, nil
}

func (kp *KafkaProducer) SendBlockMessage(block *models.BlockResponse) error {
	message := convertToBlockMessage(block)
	value, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal block message: %v", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: kp.Topic,
		Key:   sarama.StringEncoder(fmt.Sprintf("%d", message.BlockHeight)),
		Value: sarama.ByteEncoder(value),
	}

	partition, offset, err := kp.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}

	fmt.Printf("Message sent to partition %d at offset %d\n", partition, offset)
	return nil
}

func (kp *KafkaProducer) Close() error {
	return kp.producer.Close()
}

func convertToBlockMessage(block *models.BlockResponse) models.BlockMessage {
	transactions := make([]models.TransactionInfo, len(block.Result.Transactions))

	for i, tx := range block.Result.Transactions {
		transactions[i] = convertTransaction(tx)
	}

	return models.BlockMessage{
		BlockHeight:       block.Result.BlockHeight,
		BlockTime:         block.Result.BlockTime,
		Blockhash:         block.Result.Blockhash,
		ParentSlot:        block.Result.ParentSlot,
		PreviousBlockhash: block.Result.PreviousBlockhash,
		Transactions:      transactions,
		Timestamp:         time.Now().Unix(),
	}
}

func convertTransaction(tx models.Transaction) models.TransactionInfo {
	// 處理餘額變化
	balanceChanges := getBalanceChanges(tx)
	// 處理指令
	instructions := getInstructions(tx)

	status := "Success"
	if tx.Meta.Err != nil {
		status = "Failed"
	}

	return models.TransactionInfo{
		Signature:      tx.Transaction.Signatures[0],
		Status:         status,
		Fee:            tx.Meta.Fee,
		AccountKeys:    tx.Transaction.Message.AccountKeys,
		Instructions:   instructions,
		BalanceChanges: balanceChanges,
		TokenBalances:  tx.Meta.PostTokenBalances,
		ComputeUnits:   tx.Meta.ComputeUnitsConsumed,
		LogMessages:    tx.Meta.LogMessages,
	}
}

func getBalanceChanges(tx models.Transaction) []models.BalanceChange {
	changes := make([]models.BalanceChange, 0)
	for j, key := range tx.Transaction.Message.AccountKeys {
		if j < len(tx.Meta.PreBalances) && j < len(tx.Meta.PostBalances) {
			preBalance := tx.Meta.PreBalances[j]
			postBalance := tx.Meta.PostBalances[j]
			if preBalance != postBalance {
				changes = append(changes, models.BalanceChange{
					Account:     key,
					PreBalance:  preBalance,
					PostBalance: postBalance,
					Change:      int64(postBalance) - int64(preBalance),
				})
			}
		}
	}
	return changes
}

func getInstructions(tx models.Transaction) []models.Instruction {
	instructions := make([]models.Instruction, len(tx.Transaction.Message.Instructions))
	for j, inst := range tx.Transaction.Message.Instructions {
		accounts := make([]string, 0)
		for _, idx := range inst.Accounts {
			if int(idx) < len(tx.Transaction.Message.AccountKeys) {
				accounts = append(accounts, tx.Transaction.Message.AccountKeys[idx])
			}
		}

		var programId string
		if int(inst.ProgramIdIndex) < len(tx.Transaction.Message.AccountKeys) {
			programId = tx.Transaction.Message.AccountKeys[inst.ProgramIdIndex]
		}

		instructions[j] = models.Instruction{
			ProgramId: programId,
			Data:      inst.Data,
			Accounts:  accounts,
		}
	}
	return instructions
}
