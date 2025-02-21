package monitor

import (
	"fmt"
	"log"
	"time"
)

func (bm *BlockMonitor) Start() error {
	// 啟動監控報告器
	bm.startMetricsReporter()
	bm.startPendingSlotsChecker()

	// 獲取最新的 slot
	latestSlot, err := bm.solanaClient.GetLatestSlot()
	if err != nil {
		return fmt.Errorf("failed to get latest slot: %v", err)
	}

	bm.currentSlot = latestSlot
	log.Printf("Starting from slot: %d", latestSlot)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			newLatestSlot, err := bm.solanaClient.GetLatestSlot()
			if err != nil {
				log.Printf("Error getting latest slot: %v", err)
				continue
			}

			for slot := bm.currentSlot; slot <= newLatestSlot; slot++ {
				if err := bm.processBlock(slot); err != nil {
					log.Printf("Error processing block %d: %v", slot, err)
					bm.missingSlots = append(bm.missingSlots, slot)
					bm.metrics.RecordMissed()
				}
			}

			if len(bm.missingSlots) > 0 {
				bm.processMissingSlots()
			}

			bm.currentSlot = newLatestSlot

		case <-bm.stopChan:
			return nil
		}
	}
}

func (bm *BlockMonitor) processMissingSlots() {
	log.Printf("Processing %d missing blocks", len(bm.missingSlots))
	for i := 0; i < len(bm.missingSlots); i++ {
		slot := bm.missingSlots[i]
		if err := bm.processBlock(slot); err != nil {
			log.Printf("Still failed to process missing block %d: %v", slot, err)
			continue
		}
		// 成功處理後從列表中移除
		bm.missingSlots = append(bm.missingSlots[:i], bm.missingSlots[i+1:]...)
		i-- // 調整索引
	}
}
