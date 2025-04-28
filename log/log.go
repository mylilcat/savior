package log

import (
	"fmt"
	"log"
	"os"
	"time"
)

var sl *saviorLogger

type saviorLogger struct {
	logger *log.Logger
}

func NewSaviorLogger() {
	if sl != nil {
		return
	}
	sl = &saviorLogger{
		logger: log.New(os.Stdout, "[SAVIOR] ", 0),
	}
}

func Print(format string, v ...interface{}) {
	if sl == nil {
		return
	}
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, v...)

	logLine := fmt.Sprintf("%s | %s",
		timestamp,
		message,
	)
	sl.logger.Print(logLine)
}
