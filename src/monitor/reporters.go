package monitor

import (
	"fmt"
	"log"
	"strings"
	"time"
)

func (bm *BlockMonitor) startMetricsReporter() {
	bm.wg.Add(1)
	go func() {
		defer bm.wg.Done()
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				stats := bm.metrics.GetStats()
				log.Printf("Metrics Report:\n%s", formatMetrics(stats))
			case <-bm.stopChan:
				return
			}
		}
	}()
}

func (bm *BlockMonitor) startPendingSlotsChecker() {
	bm.wg.Add(1)
	go func() {
		defer bm.wg.Done()
		ticker := time.NewTicker(time.Second * 30)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				bm.checkPendingSlots()
			case <-bm.stopChan:
				return
			}
		}
	}()
}

func (bm *BlockMonitor) checkPendingSlots() {
	bm.pendingMutex.RLock()
	slots := make([]uint64, 0, len(bm.pendingSlots))
	for slot := range bm.pendingSlots {
		slots = append(slots, slot)
	}
	bm.pendingMutex.RUnlock()

	for _, slot := range slots {
		if err := bm.processBlock(slot); err != nil {
			log.Printf("Retry processing slot %d failed: %v", slot, err)
		}
	}
}

func formatMetrics(stats map[string]interface{}) string {
	var sb strings.Builder
	sb.WriteString("\n=== Block Processing Metrics ===\n")
	sb.WriteString(fmt.Sprintf("Total Processed: %d\n", stats["total_processed"]))
	sb.WriteString(fmt.Sprintf("Total Failed: %d\n", stats["total_failed"]))
	sb.WriteString(fmt.Sprintf("Total Retried: %d\n", stats["total_retried"]))
	sb.WriteString(fmt.Sprintf("Total Missed: %d\n", stats["total_missed"]))
	sb.WriteString(fmt.Sprintf("Blocks/Second: %.2f\n", stats["blocks_per_second"]))
	sb.WriteString(fmt.Sprintf("Avg Processing Time: %.2fms\n", stats["avg_processing_time"]))
	sb.WriteString(fmt.Sprintf("Success Rate: %.2f%%\n", stats["success_rate"]))
	sb.WriteString(fmt.Sprintf("Running Time: %s\n", stats["running_time"]))
	sb.WriteString(fmt.Sprintf("Last Processed Slot: %d\n", stats["last_processed_slot"]))
	sb.WriteString(fmt.Sprintf("Empty Slots: %d\n", stats["empty_slots"]))
	sb.WriteString(fmt.Sprintf("Pending Slots: %d\n", stats["pending_slots"]))
	sb.WriteString("=============================")
	return sb.String()
}
