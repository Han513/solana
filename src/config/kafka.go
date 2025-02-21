package config

import (
	"time"

	"github.com/IBM/sarama"
)

type KafkaConfig struct {
	ClientID        string
	MaxMessageBytes int
	Timeout         time.Duration
	RequiredAcks    sarama.RequiredAcks
	RetryMax        int
	Version         sarama.KafkaVersion
}

func NewKafkaConfig() *KafkaConfig {
	return &KafkaConfig{
		ClientID:        "solana-block-producer",
		MaxMessageBytes: 50 * 1024 * 1024, // 50MB
		Timeout:         60 * time.Second,
		RequiredAcks:    sarama.WaitForAll,
		RetryMax:        5,
		Version:         sarama.V3_3_1_0,
	}
}

func (c *KafkaConfig) ToSaramaConfig() *sarama.Config {
	config := sarama.NewConfig()

	config.ClientID = c.ClientID
	config.Producer.MaxMessageBytes = c.MaxMessageBytes
	config.Producer.Timeout = c.Timeout
	config.Producer.RequiredAcks = c.RequiredAcks
	config.Producer.Retry.Max = c.RetryMax
	config.Producer.Return.Successes = true
	config.Version = c.Version

	// 網絡配置
	config.Net.MaxOpenRequests = 1
	config.Net.DialTimeout = time.Second * 10
	config.Net.ReadTimeout = time.Second * 10
	config.Net.WriteTimeout = time.Second * 10

	// metadata 配置
	config.Metadata.Retry.Max = 1
	config.Metadata.Retry.Backoff = time.Second * 1
	config.Metadata.RefreshFrequency = time.Hour

	return config
}
