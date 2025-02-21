package monitor

import (
	"log"
	"time"
)

func (bm *BlockMonitor) handleEmptySlot(slot uint64) {
	bm.pendingMutex.Lock()
	defer bm.pendingMutex.Unlock()

	bm.emptySlots[slot] = &SlotStatus{
		Slot:      slot,
		Status:    "EMPTY",
		CheckTime: time.Now(),
	}

	log.Printf("Empty slot detected: %d", slot)
}

func (bm *BlockMonitor) handlePendingSlot(slot uint64, status string) {
	bm.pendingMutex.Lock()
	defer bm.pendingMutex.Unlock()

	if existing, exists := bm.pendingSlots[slot]; exists {
		existing.RetryCount++
		existing.CheckTime = time.Now()

		if existing.RetryCount >= bm.retryConfig.MaxRetries {
			bm.emptySlots[slot] = &SlotStatus{
				Slot:      slot,
				Status:    "EMPTY",
				CheckTime: time.Now(),
			}
			delete(bm.pendingSlots, slot)
			log.Printf("Slot %d marked as empty after %d retries", slot, bm.retryConfig.MaxRetries)
		}
	} else {
		bm.pendingSlots[slot] = &SlotStatus{
			Slot:       slot,
			Status:     status,
			CheckTime:  time.Now(),
			RetryCount: 1,
		}
	}
}
