package repository

import (
	"encoding/json"
	"os"

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
	// сохраняем данные в файл
	return os.WriteFile(fs.filePath, data, 0644)
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
