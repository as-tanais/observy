package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	model "github.com/as-tanais/observy/internal/model"
	"github.com/as-tanais/observy/internal/service"
)

type MetricsHandler struct {
	service *service.MetricsService
}

func NewMetricsHandler(service *service.MetricsService) *MetricsHandler {
	return &MetricsHandler{service: service}
}

func (h *MetricsHandler) UpdateMetricHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")
	metricValue := chi.URLParam(r, "value")

	if metricType == "" || metricName == "" || metricValue == "" {
		http.Error(w, "Metric type, name and value cannot be empty", http.StatusBadRequest)
		return
	}

	err := h.service.SetMetric(metricType, metricName, metricValue)
	if err != nil {
		log.Printf("Failed to set metric: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	response := fmt.Sprintf("Metric updated: type=%s, name=%s, value=%s", metricType, metricName, metricValue)
	w.Write([]byte(response))
}

func (h *MetricsHandler) GetMetricHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")

	if metricType == "" || metricName == "" {
		http.Error(w, "Metric type and name cannot be empty", http.StatusBadRequest)
		return
	}

	metric, err := h.service.GetMetric(metricName)
	if err != nil {
		log.Printf("Failed to get metric: %v", err)
		http.Error(w, "Metric not found", http.StatusNotFound)
		return
	}

	if metric.MType != metricType {
		http.Error(w, fmt.Sprintf("Metric type mismatch: expected %s, got %s", metricType, metric.MType), http.StatusBadRequest)
		return
	}

	var value string
	switch metric.MType {
	case model.Gauge:
		if metric.Value != nil {
			value = fmt.Sprintf("%g", *metric.Value)
		} else {
			http.Error(w, "Metric value is nil", http.StatusInternalServerError)
			return
		}
	case model.Counter:
		if metric.Delta != nil {
			value = fmt.Sprintf("%d", *metric.Delta)
		} else {
			http.Error(w, "Metric delta is nil", http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "Unknown metric type", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(value))
}

func (h *MetricsHandler) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	var newMetrics model.Metrics

	if err := json.NewDecoder(r.Body).Decode(&newMetrics); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	err := h.service.SetMetricWithSync(newMetrics)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)

}

func (h *MetricsHandler) GetMetric(w http.ResponseWriter, r *http.Request) {

	var metric model.Metrics

	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if metric.ID == "" {
		http.Error(w, "metric name cannot be empty", http.StatusBadRequest)
		return
	}

	output, err := h.service.GetMetric(metric.ID)
	if err != nil {
		// Если метрика не найдена - возвращаем 404
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Опционально: проверяем соответствие типа, если он указан
	if metric.MType != "" && output.MType != metric.MType {
		http.Error(w, "metric type mismatch", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(output)

}

func (h *MetricsHandler) ListMetricsHandler(w http.ResponseWriter, r *http.Request) {

	metrics := h.service.GetAllMetrics()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	html := `<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Metrics</title>
</head>
<body>
    <h1>Metrics</h1>
    <p>Total metrics: <strong>` + fmt.Sprintf("%d", len(metrics)) + `</strong></p>
`

	if len(metrics) == 0 {
		html += `<div class="empty">Метрики еще не подгружались</div>`
	} else {
		html += `
    <table>
        <thead>
            <tr>
                <th>Name</th>
                <th>Type</th>
                <th>Value</th>
            </tr>
        </thead>
        <tbody>
`
		for _, metric := range metrics {
			var value string
			var typeClass string

			switch metric.MType {
			case model.Gauge:
				typeClass = "gauge"
				if metric.Value != nil {
					value = fmt.Sprintf("%.2f", *metric.Value)
				} else {
					value = "N/A"
				}
			case model.Counter:
				typeClass = "counter"
				if metric.Delta != nil {
					value = fmt.Sprintf("%d", *metric.Delta)
				} else {
					value = "N/A"
				}
			default:
				typeClass = ""
				value = "Unknown"
			}

			html += fmt.Sprintf(`
            <tr>
                <td><strong>%s</strong></td>
                <td><span class="%s">%s</span></td>
                <td><span class="value">%s</span></td>
            </tr>
`, metric.ID, typeClass, metric.MType, value)
		}

		html += `
        </tbody>
    </table>
`
	}

	html += `
</body>
</html>
`

	w.Write([]byte(html))
}
