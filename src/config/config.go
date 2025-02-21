package config

import (
	"time"
)

type Config struct {
	RPCURL         string
	KafkaBrokers   []string
	KafkaTopic     string
	WorkerCount    int
	BatchSize      int
	MaxRetries     int
	RetryInterval  time.Duration
	MetricsEnabled bool
}

func NewConfig() *Config {
	return &Config{
		WorkerCount:    5,                // 並發處理區塊的 worker 數量
		BatchSize:      100,              // 批量處理的大小
		MaxRetries:     5,                // 最大重試次數
		RetryInterval:  time.Second * 30, // 重試間隔
		MetricsEnabled: true,             // 是否啟用 metrics
	}
}
