package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

type Logger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
	debugLogger *log.Logger
}

func NewLogger(logDir string) (*Logger, error) {
	// 確保日誌目錄存在
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %v", err)
	}

	// 創建日誌文件
	currentTime := time.Now().Format("2006-01-02")
	logFile := filepath.Join(logDir, fmt.Sprintf("solana_monitor_%s.log", currentTime))
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	// 創建不同級別的logger
	flags := log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile
	return &Logger{
		infoLogger:  log.New(file, "INFO: ", flags),
		errorLogger: log.New(file, "ERROR: ", flags),
		debugLogger: log.New(file, "DEBUG: ", flags),
	}, nil
}

func (l *Logger) Info(format string, v ...interface{}) {
	l.infoLogger.Printf(format, v...)
}

func (l *Logger) Error(format string, v ...interface{}) {
	l.errorLogger.Printf(format, v...)
}

func (l *Logger) Debug(format string, v ...interface{}) {
	l.debugLogger.Printf(format, v...)
}
