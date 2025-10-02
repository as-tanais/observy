package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/as-tanais/observy/internal/service"
)

type MetricsHandler struct {
	service *service.MetricsService
}

func NewMetricsHandler(service *service.MetricsService) *MetricsHandler {
	return &MetricsHandler{service: service}
}

func (h *MetricsHandler) UpdateMetricHandler(w http.ResponseWriter, r *http.Request) {
	method := r.Method

	if method != http.MethodPost {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 5 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	metricType := parts[2]
	metricName := parts[3]
	metricValue := parts[4]

	response := fmt.Sprintf("Type: %s, Name: %s, Value: %s", metricType, metricName, metricValue)

	err := h.service.SetMetric(metricType, metricName, metricValue)

	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	w.Write([]byte(response))

}
