package models

import (
	"sync"
	"time"
)

type Metrics struct {
	startTime          time.Time
	totalProcessed     uint64
	totalFailed        uint64
	totalRetried       uint64
	totalMissed        uint64
	processingTime     time.Duration
	lastProcessedSlot  uint64
	processedPerSecond float64
	getEmptySlots      func() int
	getPendingSlots    func() int
	mutex              sync.RWMutex
}

type MetricsStats struct {
	TotalProcessed    uint64  `json:"total_processed"`
	TotalFailed       uint64  `json:"total_failed"`
	TotalRetried      uint64  `json:"total_retried"`
	TotalMissed       uint64  `json:"total_missed"`
	BlocksPerSecond   float64 `json:"blocks_per_second"`
	AvgProcessingTime float64 `json:"avg_processing_time"`
	SuccessRate       float64 `json:"success_rate"`
	RunningTime       string  `json:"running_time"`
	LastProcessedSlot uint64  `json:"last_processed_slot"`
	EmptySlots        int     `json:"empty_slots"`
	PendingSlots      int     `json:"pending_slots"`
}

func NewMetrics(getEmptySlots, getPendingSlots func() int) *Metrics {
	return &Metrics{
		startTime:       time.Now(),
		getEmptySlots:   getEmptySlots,
		getPendingSlots: getPendingSlots,
	}
}

func (m *Metrics) UpdateMetrics(slot uint64, processTime time.Duration, isRetry bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.totalProcessed++
	m.processingTime += processTime
	m.lastProcessedSlot = slot

	if isRetry {
		m.totalRetried++
	}

	elapsed := time.Since(m.startTime).Seconds()
	m.processedPerSecond = float64(m.totalProcessed) / elapsed
}

// RecordFailure 記錄失敗
func (m *Metrics) RecordFailure() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.totalFailed++
}

// RecordMissed 記錄遺漏
func (m *Metrics) RecordMissed() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.totalMissed++
}

// GetStats 獲取統計信息
func (m *Metrics) GetStats() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var avgProcessingTime float64
	var successRate float64

	if m.totalProcessed > 0 {
		avgProcessingTime = float64(m.processingTime.Milliseconds()) / float64(m.totalProcessed)
		successRate = float64(m.totalProcessed-m.totalFailed) / float64(m.totalProcessed) * 100
	}

	return map[string]interface{}{
		"total_processed":     m.totalProcessed,
		"total_failed":        m.totalFailed,
		"total_retried":       m.totalRetried,
		"total_missed":        m.totalMissed,
		"blocks_per_second":   m.processedPerSecond,
		"avg_processing_time": avgProcessingTime,
		"success_rate":        successRate,
		"running_time":        time.Since(m.startTime).String(),
		"last_processed_slot": m.lastProcessedSlot,
		"empty_slots":         m.getEmptySlots(),
		"pending_slots":       m.getPendingSlots(),
	}
}
