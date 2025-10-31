package agent

import (
	"math/rand"
	"runtime"

	models "github.com/as-tanais/observy/internal/model"
)

func toFloat64(v uint64) *float64 {
	f := float64(v)
	return &f
}

func toFloat64FromUint32(v uint32) *float64 {
	f := float64(v)
	return &f
}

func Collect() []models.Metrics {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	delta := int64(1)

	randomValue := rand.Float64()

	metrics := make([]models.Metrics, 0, 29)

	metrics = append(metrics,
		models.Metrics{ID: "Alloc", MType: models.Gauge, Value: toFloat64(memStats.Alloc)},
		models.Metrics{ID: "BuckHashSys", MType: models.Gauge, Value: toFloat64(memStats.BuckHashSys)},
		models.Metrics{ID: "Frees", MType: models.Gauge, Value: toFloat64(memStats.Frees)},
		models.Metrics{ID: "GCCPUFraction", MType: models.Gauge, Value: &memStats.GCCPUFraction},
		models.Metrics{ID: "GCSys", MType: models.Gauge, Value: toFloat64(memStats.GCSys)},
		models.Metrics{ID: "HeapAlloc", MType: models.Gauge, Value: toFloat64(memStats.HeapAlloc)},
		models.Metrics{ID: "HeapIdle", MType: models.Gauge, Value: toFloat64(memStats.HeapIdle)},
		models.Metrics{ID: "HeapInuse", MType: models.Gauge, Value: toFloat64(memStats.HeapInuse)},
		models.Metrics{ID: "HeapObjects", MType: models.Gauge, Value: toFloat64(memStats.HeapObjects)},
		models.Metrics{ID: "HeapReleased", MType: models.Gauge, Value: toFloat64(memStats.HeapReleased)},
		models.Metrics{ID: "HeapSys", MType: models.Gauge, Value: toFloat64(memStats.HeapSys)},
		models.Metrics{ID: "LastGC", MType: models.Gauge, Value: toFloat64(memStats.LastGC)},
		models.Metrics{ID: "Lookups", MType: models.Gauge, Value: toFloat64(memStats.Lookups)},
		models.Metrics{ID: "MCacheInuse", MType: models.Gauge, Value: toFloat64(memStats.MCacheInuse)},
		models.Metrics{ID: "MCacheSys", MType: models.Gauge, Value: toFloat64(memStats.MCacheSys)},
		models.Metrics{ID: "MSpanInuse", MType: models.Gauge, Value: toFloat64(memStats.MSpanInuse)},
		models.Metrics{ID: "MSpanSys", MType: models.Gauge, Value: toFloat64(memStats.MSpanSys)},
		models.Metrics{ID: "Mallocs", MType: models.Gauge, Value: toFloat64(memStats.Mallocs)},
		models.Metrics{ID: "NextGC", MType: models.Gauge, Value: toFloat64(memStats.NextGC)},
		models.Metrics{ID: "NumForcedGC", MType: models.Gauge, Value: toFloat64FromUint32(memStats.NumForcedGC)},
		models.Metrics{ID: "NumGC", MType: models.Gauge, Value: toFloat64FromUint32(memStats.NumGC)},
		models.Metrics{ID: "OtherSys", MType: models.Gauge, Value: toFloat64(memStats.OtherSys)},
		models.Metrics{ID: "PauseTotalNs", MType: models.Gauge, Value: toFloat64(memStats.PauseTotalNs)},
		models.Metrics{ID: "StackInuse", MType: models.Gauge, Value: toFloat64(memStats.StackInuse)},
		models.Metrics{ID: "StackSys", MType: models.Gauge, Value: toFloat64(memStats.StackSys)},
		models.Metrics{ID: "Sys", MType: models.Gauge, Value: toFloat64(memStats.Sys)},
		models.Metrics{ID: "TotalAlloc", MType: models.Gauge, Value: toFloat64(memStats.TotalAlloc)},

		models.Metrics{ID: "PollCount", MType: models.Counter, Delta: &delta},
		models.Metrics{ID: "RandomValue", MType: models.Gauge, Value: &randomValue},
	)

	return metrics
}
