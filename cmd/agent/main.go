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
	pollInterval := flag.Duration("p", 2*time.Second, "Poll interval, default: 2s")
	reportInterval := flag.Duration("r", 10*time.Second, "Report interval, default: 10s")

	flag.Parse()

	serverURL := "http://" + *serverAddr

	pollsPerReport := int(*reportInterval / *pollInterval)

	fmt.Printf("Starting agent: server=%s, poll=%v, report=%v",
		serverURL, *pollInterval, *reportInterval)

	for {
		var metrics []models.Metrics

		for i := 0; i < pollsPerReport; i++ {
			metrics = agent.Collect()

			if i < pollsPerReport-1 {
				time.Sleep(*pollInterval)
			}
		}
		agent.Send(metrics, serverURL)

		time.Sleep(*pollInterval)

	}
}
