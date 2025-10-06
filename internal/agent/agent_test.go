package agent

import (
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollect(t *testing.T) {

	metrics := Collect()

	require.NotEmpty(t, metrics, "Метрики не должны быть пустыми")

	expectedCount := 29 //Все метрики из задания(библиотека runtime) + 2 доп
	assert.Equal(t, expectedCount, len(metrics), "Должно быть собрано 29 метрик")

	for _, metric := range metrics {
		assert.NotEmpty(t, metric.ID, "ID метрики не должен быть пустым")
		assert.NotEmpty(t, metric.MType, "Тип метрики не должен быть пустым")
	}
}

func TestCollect_PollCountIncrement(t *testing.T) {
	atomic.StoreInt64(&pollCount, 0)

	for i := 1; i <= 3; i++ {
		metrics := Collect()

		var found bool
		var pollCountValue int64

		for _, metric := range metrics {
			if metric.ID == "PollCount" {
				found = true
				require.NotNil(t, metric.Delta, "PollCount Delta не должна быть nil")
				pollCountValue = *metric.Delta
				break
			}
		}

		assert.True(t, found, "Метрика PollCount должна присутствовать")

		assert.Equal(t, int64(i), pollCountValue, "PollCount должен быть равен %d после %d вызова", i, i)
	}
}
