package repository

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	models "github.com/as-tanais/observy/internal/model"
)

type FileStorage struct {
	mu       sync.RWMutex
	metrics  map[string]models.Metrics
	filePath string
}

func NewFileStorage(filePath string) *FileStorage {
	return &FileStorage{
		filePath: filePath,
		metrics:  make(map[string]models.Metrics),
	}
}

func (fs *FileStorage) SaveMetrics(metrics []models.Metrics) error {
	data, err := json.MarshalIndent(metrics, "", "   ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(fs.filePath)
	fileName := filepath.Base(fs.filePath)
	tmpFile := filepath.Join(dir, "."+fileName+".tmp")

	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return err
	}

	return os.Rename(tmpFile, fs.filePath)
}

func (fs *FileStorage) LoadMetrics() ([]models.Metrics, error) {
	data, err := os.ReadFile(fs.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.Metrics{}, nil
		}
		return nil, err
	}

	var metrics []models.Metrics

	if err := json.Unmarshal(data, &metrics); err != nil {
		return nil, err
	}
	return metrics, nil
}

func (fs *FileStorage) SetMetric(_ context.Context, m *models.Metrics) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.metrics[m.ID] = *m

	return nil
}

func (fs *FileStorage) GetMetric(_ context.Context, id string) (models.Metrics, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	m, ok := fs.metrics[id]

	if !ok {
		return models.Metrics{}, ErrMetricNotFound
	}

	return m, nil
}

func (fs *FileStorage) GetAllMetrics(_ context.Context) ([]models.Metrics, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	metrics := make([]models.Metrics, 0, len(fs.metrics))
	for _, metric := range fs.metrics {
		metrics = append(metrics, metric)
	}
	return metrics, nil
}
