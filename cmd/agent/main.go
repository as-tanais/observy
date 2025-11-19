package main

import (
	"fmt"
	"log"
	"time"

	"github.com/as-tanais/observy/internal/agent"
	"github.com/as-tanais/observy/internal/config"
	models "github.com/as-tanais/observy/internal/model"
)

func main() {
	cfg, err := config.NewAgentConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Starting agent: server=%s, poll=%v, report=%v\n",
		cfg.ServerURL(), cfg.PollInterval, cfg.ReportInterval)

	for {
		var metrics []models.Metrics

		for i := 0; i < cfg.PollsPerReport(); i++ {
			metrics = agent.Collect()

			if i < cfg.PollsPerReport()-1 {
				time.Sleep(cfg.PollInterval)
			}
		}

		agent.Send(metrics, cfg.ServerURL(), cfg.Key)
		agent.SendBatchMetrics(metrics, cfg.ServerURL(), cfg.Key)

		time.Sleep(cfg.PollInterval)
	}
}
