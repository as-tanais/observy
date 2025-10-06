package agent

import (
	"fmt"
	"log"
	"net/http"

	models "github.com/as-tanais/observy/internal/model"
)

const serverAddress = "http://localhost:8080"

func Send(metrics []models.Metrics) {
	client := &http.Client{}

	for _, metric := range metrics {
		var value string

		// Определяем значение в зависимости от типа метрики
		switch metric.MType {
		case models.Gauge:
			if metric.Value != nil {
				value = fmt.Sprintf("%f", *metric.Value)
			} else {
				log.Printf("Пропущена gauge метрика %s: значение nil", metric.ID)
				continue
			}
		case models.Counter:
			if metric.Delta != nil {
				value = fmt.Sprintf("%d", *metric.Delta)
			} else {
				log.Printf("Пропущена counter метрика %s: значение nil", metric.ID)
				continue
			}
		default:
			log.Printf("Неизвестный тип метрики: %s", metric.MType)
			continue
		}

		// Формируем URL
		url := fmt.Sprintf("%s/update/%s/%s/%s", serverAddress, metric.MType, metric.ID, value)

		// Создаём запрос
		req, err := http.NewRequest(http.MethodPost, url, nil)
		if err != nil {
			log.Printf("Ошибка создания запроса для %s: %v", metric.ID, err)
			continue
		}

		// Устанавливаем заголовок
		req.Header.Set("Content-Type", "text/plain")

		// Отправляем запрос
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Ошибка отправки метрики %s: %v", metric.ID, err)
			continue
		}

		// Закрываем тело ответа
		resp.Body.Close()

		// Проверяем статус ответа
		if resp.StatusCode != http.StatusOK {
			log.Printf("Метрика %s: неожиданный статус %d", metric.ID, resp.StatusCode)
		}
	}
}
