package handler

import (
	"fmt"
	"log"
	"net/http"
	"strings"

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

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	if len(parts) != 4 || parts[0] != "update" {
		http.Error(w, "Invalid path format. Expected: /update/{type}/{name}/{value}", http.StatusNotFound)
		return
	}

	metricType := parts[1]
	metricName := parts[2]
	metricValue := parts[3]

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

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	if len(parts) != 3 || parts[0] != "value" {
		http.Error(w, "Invalid path format. Expected: /value/{type}/{name}", http.StatusNotFound)
		return
	}

	metricType := parts[1]
	metricName := parts[2]

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
