package repository

import (
	"encoding/json"
	"os"
	"path/filepath"

	models "github.com/as-tanais/observy/internal/model"
)

type FileStorage struct {
	filePath string
}

func NewFileStorage(filePath string) *FileStorage {
	return &FileStorage{filePath: filePath}
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
