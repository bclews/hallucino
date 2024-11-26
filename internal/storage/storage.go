package storage

import (
	"fmt"
	"hallucino/internal/k8s"
	"sync"

	"github.com/fatih/color"
)

type LogStorage struct {
	logs []k8s.LogEntry
	mu   sync.RWMutex
}

func NewLogStorage() *LogStorage {
	return &LogStorage{
		logs: []k8s.LogEntry{},
	}
}

func (ls *LogStorage) AddLog(log k8s.LogEntry) {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	ls.logs = append(ls.logs, log)
}

func (ls *LogStorage) GetLogs() []k8s.LogEntry {
	ls.mu.RLock()
	defer ls.mu.RUnlock()
	return ls.logs
}

func (ls *LogStorage) PrettyPrintLogs() {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	// Use different colors for different elements
	podColor := color.New(color.FgBlue).SprintFunc()
	containerColor := color.New(color.FgMagenta).SprintFunc()
	timestampColor := color.New(color.FgGreen).SprintFunc()

	for _, log := range ls.logs {
		// Format log entry
		fmt.Printf("%s | %s | %s | %s\n",
			timestampColor(log.Timestamp),
			podColor(log.PodName),
			containerColor(log.Container),
			log.LogContent,
		)
	}
}

func (ls *LogStorage) Clear() {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	ls.logs = []k8s.LogEntry{}
}
