package main

import (
	"context"
	"crypto/rsa"
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
	"github.com/as-tanais/observy/internal/crypto"
	models "github.com/as-tanais/observy/internal/model"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func worker(jobs <-chan []models.Metrics,
	address, key string,
	pubKey *rsa.PublicKey,
	grpcClient *agent.GRPCClient,
	wg *sync.WaitGroup,
	reportInterval time.Duration) {

	defer wg.Done()

	for m := range jobs {

		if grpcClient != nil {
			ctx := context.Background()
			if err := grpcClient.SendMetricsBatch(ctx, m); err == nil {

				time.Sleep(reportInterval)
				continue
			} else {
				log.Printf("gRPC send failed, falling back to HTTP: %v", err)
			}
		}

		agent.Send(m, address, key, pubKey)
		agent.SendBatchMetrics(m, address, key, pubKey)

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

	var grpcClient *agent.GRPCClient
	if cfg.GRPCAddress != "" {
		log.Printf("Initializing gRPC client to %s", cfg.GRPCAddress)
		grpcClient, err = agent.NewGRPCClient(cfg.GRPCAddress)
		if err != nil {
			log.Printf("Failed to create gRPC client go work with HTTP: %v", err)
			grpcClient = nil
		} else {

			grpcClient = grpcClient.WithTimeout(15 * time.Second)
			defer grpcClient.Close()
			log.Println("gRPC client initialized successfully")
		}
	}

	tasks := make(chan []models.Metrics, 100)

	var wg sync.WaitGroup
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	ctx, cancel := context.WithCancel(context.Background())

	totalGoR := cfg.RateLimit + 3
	wg.Add(totalGoR)

	var publicKey *rsa.PublicKey
	if cfg.CryptoKey != "" {
		log.Printf("Loading public key from: %s", cfg.CryptoKey)
		key, err := crypto.LoadPublicKey(cfg.CryptoKey)
		if err != nil {
			log.Fatalf("Failed to load public key: %v", err)
		}
		publicKey = key
	}

	for i := 0; i < cfg.RateLimit; i++ {
		go worker(tasks, cfg.ServerURL(), cfg.Key, publicKey, grpcClient, &wg, cfg.ReportInterval)
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
	log.Println("Shutting down agent")
	cancel()
	time.Sleep(1 * time.Second)
	close(tasks)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("All workers stoped")
	case <-time.After(5 * time.Second):
		log.Println("Wait workers")
	}

	log.Println("Agent stopped")
}
