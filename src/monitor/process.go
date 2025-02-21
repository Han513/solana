package monitor

import (
	"fmt"
	"log"
	"time"
)

func (bm *BlockMonitor) processBlock(slot uint64) error {
	// 檢查是否已處理
	if bm.processedSlots[slot] {
		return nil
	}

	// 檢查區塊狀態
	status, err := bm.solanaClient.CheckSlotStatus(slot)
	if err != nil {
		bm.handlePendingSlot(slot, "NOT_AVAILABLE")
		return fmt.Errorf("failed to check slot status: %v", err)
	}

	switch status {
	case "EMPTY":
		bm.handleEmptySlot(slot)
		return nil
	case "NOT_AVAILABLE":
		bm.handlePendingSlot(slot, status)
		return fmt.Errorf("block not available yet")
	case "CONFIRMED":
		return bm.processConfirmedBlock(slot)
	}

	return nil
}

func (bm *BlockMonitor) processConfirmedBlock(slot uint64) error {
	startTime := time.Now()
	var lastErr error

	// 重試邏輯
	for retry := 0; retry <= bm.retryConfig.MaxRetries; retry++ {
		if retry > 0 {
			delay := bm.retryConfig.Backoff(retry)
			time.Sleep(delay)
		}

		// 獲取區塊數據
		block, err := bm.solanaClient.GetBlock(slot)
		if err != nil {
			lastErr = fmt.Errorf("attempt %d: failed to get block details: %v", retry+1, err)
			continue
		}

		// 發送到 Kafka
		if err := bm.producer.SendBlockMessage(block); err != nil {
			lastErr = fmt.Errorf("attempt %d: failed to send to kafka: %v", retry+1, err)
			continue
		}

		// 處理成功
		processTime := time.Since(startTime)
		bm.metrics.UpdateMetrics(slot, processTime, retry > 0)

		// 更新狀態
		bm.pendingMutex.Lock()
		delete(bm.pendingSlots, slot)
		bm.processedSlots[slot] = true
		bm.pendingMutex.Unlock()

		log.Printf("Successfully processed block %d (retry: %d, time: %v)", slot, retry, processTime)
		return nil
	}

	bm.metrics.RecordFailure()
	return lastErr
}
