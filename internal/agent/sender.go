package agent

import (
	"fmt"
	"log"
	"net/http"
	"time"

	models "github.com/as-tanais/observy/internal/model"
)

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

func Send(metrics []models.Metrics, serverAddress string) {

	for _, metric := range metrics {
		var value string

		// Определяем значение в зависимости от типа метрики
		switch metric.MType {
		case models.Gauge:
			if metric.Value != nil {
				value = fmt.Sprintf("%f", *metric.Value)
			} else {
				log.Printf("Пропущена gauge метрика %s: значение nil\n", metric.ID)
				continue
			}
		case models.Counter:
			if metric.Delta != nil {
				value = fmt.Sprintf("%d", *metric.Delta)
			} else {
				log.Printf("Пропущена counter метрика %s: значение nil\n", metric.ID)
				continue
			}
		default:
			log.Printf("Неизвестный тип метрики: %s\n", metric.MType)
			continue
		}

		// Формируем URL
		url := fmt.Sprintf("%s/update/%s/%s/%s", serverAddress, metric.MType, metric.ID, value)

		// Создаём запрос
		req, err := http.NewRequest(http.MethodPost, url, nil)
		if err != nil {
			log.Printf("Ошибка создания запроса для %s: %v\n", metric.ID, err)
			continue
		}

		// Устанавливаем заголовок
		req.Header.Set("Content-Type", "text/plain")

		// Отправляем запрос
		resp, err := httpClient.Do(req)
		if err != nil {
			log.Printf("Ошибка отправки метрики %s: %v\n", metric.ID, err)
			continue
		}

		// Закрываем тело ответа
		resp.Body.Close()

		// Проверяем статус ответа
		if resp.StatusCode != http.StatusOK {
			log.Printf("Метрика %s: неожиданный статус %d\ns", metric.ID, resp.StatusCode)
		}
	}
}
