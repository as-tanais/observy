package main

import (
	"time"

	"github.com/as-tanais/observy/internal/agent"
	models "github.com/as-tanais/observy/internal/model"
)

func main() {
	pollInterval := 2 * time.Second
	reportInterval := 10 * time.Second

	pollsPerReport := int(reportInterval / pollInterval)

	for {
		var metrics []models.Metrics

		for i := 0; i < pollsPerReport; i++ {
			metrics = agent.Collect()

			if i < pollsPerReport-1 {
				time.Sleep(pollInterval)
			}
		}
		agent.Send(metrics)

		time.Sleep(pollInterval)

	}
}
