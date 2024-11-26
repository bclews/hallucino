package storage

import (
	"hallucino/internal/k8s"
	"sync"
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

func (ls *LogStorage) Clear() {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	ls.logs = []k8s.LogEntry{}
}
