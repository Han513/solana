package monitor

import (
	"sync"

	"solana/src/config"
	"solana/src/models"
	"solana/src/services"
)

type BlockMonitor struct {
	rpcURL         string
	producer       *services.KafkaProducer
	topic          string
	currentSlot    uint64
	processedSlots map[uint64]bool
	missingSlots   []uint64
	metrics        *models.Metrics
	retryConfig    *config.RetryConfig
	stopChan       chan struct{}
	wg             sync.WaitGroup
	emptySlots     map[uint64]*SlotStatus // 記錄空槽
	pendingSlots   map[uint64]*SlotStatus // 記錄未確認的區塊
	pendingMutex   sync.RWMutex           // 保護 maps 的互斥鎖
	solanaClient   *services.SolanaClient
}

func NewBlockMonitor(rpcURL string, producer *services.KafkaProducer, topic string) *BlockMonitor {
	bm := &BlockMonitor{
		rpcURL:         rpcURL,
		producer:       producer,
		topic:          topic,
		processedSlots: make(map[uint64]bool),
		missingSlots:   make([]uint64, 0),
		emptySlots:     make(map[uint64]*SlotStatus),
		pendingSlots:   make(map[uint64]*SlotStatus),
		retryConfig:    config.NewRetryConfig(),
		stopChan:       make(chan struct{}),
		solanaClient:   services.NewSolanaClient(rpcURL),
	}

	// 初始化 metrics
	bm.metrics = models.NewMetrics(
		func() int {
			bm.pendingMutex.RLock()
			defer bm.pendingMutex.RUnlock()
			return len(bm.emptySlots)
		},
		func() int {
			bm.pendingMutex.RLock()
			defer bm.pendingMutex.RUnlock()
			return len(bm.pendingSlots)
		},
	)

	return bm
}
