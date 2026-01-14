package audit

import (
	"context"
	"encoding/json"
	"os"
	"sync"
)

type FileSub struct {
	filePath string
	file     *os.File
	mu       sync.Mutex
}

func NewFileSub(filePath string) (*FileSub, error) {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &FileSub{
		filePath: filePath,
		file:     file,
	}, nil
}

func (f *FileSub) Send(ctx context.Context, event AuditEvent) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	data = append(data, '\n')
	_, err = f.file.Write(data)
	return err
}
