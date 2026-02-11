package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/as-tanais/observy/internal/agent"
	"github.com/as-tanais/observy/internal/buildinfo"
	"github.com/as-tanais/observy/internal/config"
	models "github.com/as-tanais/observy/internal/model"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func worker(jobs <-chan []models.Metrics, address, key string, wg *sync.WaitGroup, reportInterval time.Duration) {
	defer wg.Done()
	for m := range jobs {
		agent.Send(m, address, key)
		agent.SendBatchMetrics(m, address, key)
		time.Sleep(reportInterval)
	}
}

func main() {
	buildinfo.PrintInfo(buildVersion, buildDate, buildCommit)

	cfg, err := config.NewAgentConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Starting agent: server=%s, poll=%v, report=%v\n",
		cfg.ServerURL(), cfg.PollInterval, cfg.ReportInterval)

	tasks := make(chan []models.Metrics, 100)

	var wg sync.WaitGroup
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())

	totalGoR := cfg.RateLimit + 3

	wg.Add(totalGoR)

	for i := 0; i < cfg.RateLimit; i++ {
		go worker(tasks, cfg.ServerURL(), cfg.Key, &wg, cfg.ReportInterval)
	}

	go func() {

		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:

				metrics := agent.Collect()
				tasks <- metrics

				time.Sleep(cfg.PollInterval)

			}
		}

	}()

	cpuDataChan := make(chan []float64, 1)

	go func() {
		defer wg.Done()
		agent.CollectCPUData(cpuDataChan, ctx)
	}()

	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():

				return
			default:

				metrics := agent.CollectSystemMetrics(cpuDataChan)
				tasks <- metrics

				time.Sleep(cfg.PollInterval)

			}

		}

	}()

	<-sigChan
	cancel()
	close(tasks)

	wg.Wait()

}
