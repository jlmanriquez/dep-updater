package trace

import (
	"fmt"
	"log"
	"os"
	"time"
)

type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	ErrLevel
)

var (
	currentLevel = InfoLevel

	fileLog    *log.Logger
	consoleLog *log.Logger

	logPrefix = map[Level]string{
		DebugLevel: "DEB",
		InfoLevel:  "INF",
		ErrLevel:   "ERR",
	}
)

func Init(l Level) {
	currentLevel = l

	now := time.Now()

	logTime := now.Format("20060102150405")
	file, err := os.OpenFile(fmt.Sprintf("updater_%s.log", logTime), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(fmt.Sprintf("error creating log file...%s", err.Error()))
	}

	fileLog = log.New(file, "", log.LstdFlags)
	consoleLog = log.New(os.Stdout, "", log.LstdFlags)
}

func Trace(level Level, prefix, msg string, logs ...*log.Logger) {
	if level >= currentLevel {
		var message string

		if prefix != "" {
			message = fmt.Sprintf("[%s][%s] %s", logPrefix[level], prefix, msg)
		} else {
			message = fmt.Sprintf("[%s] %s", logPrefix[level], msg)
		}

		for _, l := range logs {
			l.Printf(message)
		}
	}
}

func Debugp(prefix, msg string) {
	Trace(DebugLevel, prefix, msg, fileLog)
}

func Infop(prefix, msg string) {
	Trace(InfoLevel, prefix, msg, fileLog)
}

func Info(msg string) {
	Infop("", msg)
}

func InfoConsole(prefix, msg string) {
	Trace(InfoLevel, prefix, msg, consoleLog, fileLog)
}

func ErrorConsole(prefix, msg string) {
	Trace(ErrLevel, prefix, msg, consoleLog, fileLog)
}
