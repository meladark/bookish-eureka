package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

type Level int

const (
	ErrorLevel Level = iota
	WarnLevel
	InfoLevel
	DebugLevel
)

func parseLevel(levelStr string) Level {
	switch strings.ToLower(levelStr) {
	case "error":
		return ErrorLevel
	case "warn":
		return WarnLevel
	case "info":
		return InfoLevel
	case "debug":
		return DebugLevel
	default:
		return InfoLevel
	}
}

type Logger struct {
	level   Level
	std     *log.Logger
	logFile *os.File
}

func New(level string, logFilePath string) *Logger {
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		log.Fatalf("cannot open log file: %v", err)
	}

	multiWriter := io.MultiWriter(os.Stdout, logFile)

	return &Logger{
		level:   parseLevel(level),
		std:     log.New(multiWriter, "", 0),
		logFile: logFile,
	}
}

func (l *Logger) Close() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

func (l *Logger) SetLevel(level string) Level {
	if l.level != parseLevel(level) {
		l.level = parseLevel(level)
	}
	return l.level
}

func (l *Logger) log(level Level, msg string) {
	if level > l.level {
		return
	}

	levelStr := ""
	switch level {
	case ErrorLevel:
		levelStr = "ERROR"
	case WarnLevel:
		levelStr = "WARN"
	case InfoLevel:
		levelStr = "INFO"
	case DebugLevel:
		levelStr = "DEBUG"
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	output := fmt.Sprintf("%s [%s] %s", timestamp, levelStr, msg)
	l.std.Println(output)
}

func (l *Logger) Error(msg string) {
	l.log(ErrorLevel, msg)
}

func (l *Logger) Warn(msg string) {
	l.log(WarnLevel, msg)
}

func (l *Logger) Info(msg string) {
	l.log(InfoLevel, msg)
}

func (l *Logger) Debug(msg string) {
	l.log(DebugLevel, msg)
}
