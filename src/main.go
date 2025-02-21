package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"solana/src/config"
	"solana/src/monitor"
	"solana/src/services"
	"solana/src/utils"

	"github.com/joho/godotenv"
)

var (
	configFile = flag.String("config", ".env", "Path to configuration file")
	logDir     = flag.String("log-dir", "logs", "Directory for log files")
	version    = flag.Bool("version", false, "Show version information")
)

func main() {
	flag.Parse()

	// 顯示版本信息
	if *version {
		fmt.Println(utils.GetVersionInfo())
		return
	}

	// 初始化日誌
	logger, err := utils.NewLogger(*logDir)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// 加載配置
	if err := godotenv.Load("../.env"); err != nil {
		logger.Error("Error loading config file: %v", err)
		os.Exit(1)
	}

	// 獲取配置值
	rpcURL := os.Getenv("RPC_URL")
	if rpcURL == "" {
		logger.Error("RPC_URL not set in config file")
		os.Exit(1)
	}

	// 測試連接
	checker := utils.NewTCPConnectionChecker(5 * time.Second)
	if err := checker.TestConnection("127.0.0.1", "8998"); err != nil {
		logger.Error("Kafka connection test failed: %v", err)
		// 在這裡可以選擇是否退出
		// os.Exit(1)
	} else {
		logger.Info("Kafka connection test successful")
	}

	// 創建 Kafka 生產者
	kafkaConfig := config.NewKafkaConfig()
	producer, err := services.NewKafkaProducer(
		kafkaConfig,
		[]string{"127.0.0.1:8998"},
	)
	if err != nil {
		logger.Error("Failed to create Kafka producer: %v", err)
		os.Exit(1)
	}

	// 設置 topic
	producer.Topic = "solana"

	defer producer.Close()

	// 創建並啟動監視器
	monitor := monitor.NewBlockMonitor(rpcURL, producer, "solana")
	defer monitor.Stop()

	// 處理系統信號
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 啟動監視器
	go func() {
		if err := monitor.Start(); err != nil {
			logger.Error("Monitor failed: %v", err)
			os.Exit(1)
		}
	}()

	logger.Info("Solana block monitor started")

	// 等待退出信號
	sig := <-sigChan
	logger.Info("Received signal %v, shutting down...", sig)

	// 優雅關閉
	monitor.Stop()
	logger.Info("Shutdown complete")
}
