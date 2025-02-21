package main

import (
	"fmt"
	"log"
	"time"

	"github.com/IBM/sarama"
)

func createKafkaProducer() (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	// 設置版本
	config.Version = sarama.V2_8_0_0

	// 設置超時時間
	config.Net.DialTimeout = time.Second * 30
	config.Net.ReadTimeout = time.Second * 30
	config.Net.WriteTimeout = time.Second * 30

	producer, err := sarama.NewSyncProducer([]string{"127.0.0.1:8998"}, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %v", err)
	}

	return producer, nil
}

func main() {
	// 創建生產者
	producer, err := createKafkaProducer()
	if err != nil {
		log.Fatalf("Failed to create producer: %v", err)
	}
	defer producer.Close()

	// 創建消息
	msg := &sarama.ProducerMessage{
		Topic: "solana",
		Key:   sarama.StringEncoder("test-key"),
		Value: sarama.StringEncoder("test2"),
	}

	// 發送消息
	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		log.Fatalf("Failed to send message: %v", err)
	}

	fmt.Printf("Message sent successfully! Partition: %d, Offset: %d\n", partition, offset)
}
