package logger

import (
	"bufio"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const expected = `[INFO] Info message
[INFO] info message Info
[WARN] warn message Info
[ERROR] error message Info
[DEBUG] Debug message
[INFO] info message Debug
[DEBUG] debug message Debug
[WARN] warn message Debug
[ERROR] error message Debug
[WARN] Warn message
[WARN] warn message Warn
[ERROR] error message Warn
[ERROR] Error message
[ERROR] error message Error`

func TestLoggerLevels(t *testing.T) {
	patest := "./test.log"
	log := New("info", patest)
	l := "Empty"
	for i := range 4 {
		switch i {
		case 0:
			log.SetLevel("info")
			log.Info("Info message")
			l = "Info"
		case 1:
			log.SetLevel("debug")
			log.Debug("Debug message")
			l = "Debug"
		case 2:
			log.SetLevel("warn")
			log.Warn("Warn message")
			l = "Warn"
		case 3:
			log.SetLevel("error")
			log.Error("Error message")
			l = "Error"
		}
		log.Info("info message " + l)
		log.Debug("debug message " + l)
		log.Warn("warn message " + l)
		log.Error("error message " + l)
	}
	log.Close()

	file, err := os.Open(patest)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 20 {
			line = line[20:]
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("error reading file: %v", err)
	}

	got := strings.Join(lines, "\n")

	require.Equal(t, expected, got)
	os.Remove(patest)
}
