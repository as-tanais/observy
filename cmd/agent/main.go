package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/as-tanais/observy/internal/agent"
	models "github.com/as-tanais/observy/internal/model"
)

func main() {

	serverAddr := flag.String("a", "localhost:8080", "Server address host:port, default: localhost:8080")
	pollIntervalSec := flag.Int("p", 2, "Poll interval, default: 2s")
	reportIntervalSec := flag.Int("r", 10, "Report interval, default: 10s")

	flag.Parse()

	serverURL := "http://" + *serverAddr
	pollInterval := time.Duration(*pollIntervalSec) * time.Second
	reportInterval := time.Duration(*reportIntervalSec) * time.Second

	pollsPerReport := *reportIntervalSec / *pollIntervalSec

	fmt.Printf("Starting agent: server=%s, poll=%v, report=%v",
		serverURL, pollInterval, reportInterval)

	for {
		var metrics []models.Metrics

		for i := 0; i < pollsPerReport; i++ {
			metrics = agent.Collect()

			if i < pollsPerReport-1 {
				time.Sleep(pollInterval)
			}
		}
		agent.Send(metrics, serverURL)

		time.Sleep(pollInterval)

	}
}
