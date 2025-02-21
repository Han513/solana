package config

import (
	"time"
)

type RetryConfig struct {
	MaxRetries    int
	InitialDelay  time.Duration
	MaxDelay      time.Duration
	BackoffFactor float64
}

func NewRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:    5,
		InitialDelay:  time.Second,
		MaxDelay:      time.Second * 30,
		BackoffFactor: 2.0,
	}
}

// Backoff 計算下一次重試的延遲時間
func (c *RetryConfig) Backoff(attempt int) time.Duration {
	if attempt <= 0 {
		return c.InitialDelay
	}

	delay := c.InitialDelay
	for i := 0; i < attempt; i++ {
		delay = time.Duration(float64(delay) * c.BackoffFactor)
		if delay > c.MaxDelay {
			return c.MaxDelay
		}
	}
	return delay
}
