package monitor

import "time"

type SlotStatus struct {
	Slot       uint64
	Status     string // "EMPTY", "NOT_AVAILABLE", "CONFIRMED"
	CheckTime  time.Time
	RetryCount int
}
