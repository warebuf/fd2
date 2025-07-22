package logger

import (
	"os"
	"path/filepath"
	"sync"
)

type Logger struct {
	file  *os.File
	mutex sync.Mutex
}

func NewLogger(filename string) (*Logger, error) {
	logDir := "logs"
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		return nil, err
	}

	fullPath := filepath.Join(logDir, filename)
	f, err := os.OpenFile(fullPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &Logger{file: f}, nil
}

func (l *Logger) Write(line string) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	_, err := l.file.WriteString(line + "\n")
	return err
}

func (l *Logger) Close() error {
	return l.file.Close()
}
