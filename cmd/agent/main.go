package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/as-tanais/observy/internal/agent"
	models "github.com/as-tanais/observy/internal/model"
)

func main() {

	defaultServerAddr := "localhost:8080"
	defaultPollIntervalSec := 2
	defaultReportIntervalSec := 10

	addrFlag := flag.String("a", defaultServerAddr, "Server address host:port")
	pollFlag := flag.Int("p", defaultPollIntervalSec, "Poll interval in seconds")
	reportFlag := flag.Int("r", defaultReportIntervalSec, "Report interval in seconds")

	flag.Parse()

	serverAddr := *addrFlag
	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		serverAddr = envAddr
	}

	pollIntervalSec := *pollFlag
	if envPoll := os.Getenv("POLL_INTERVAL"); envPoll != "" {
		if val, err := strconv.Atoi(envPoll); err == nil {
			pollIntervalSec = val
		} else {
			fmt.Fprintf(os.Stderr, "Invalid POLL_INTERVAL: %v\n", err)
			os.Exit(1)
		}
	}

	reportIntervalSec := *reportFlag
	if envReport := os.Getenv("REPORT_INTERVAL"); envReport != "" {
		if val, err := strconv.Atoi(envReport); err == nil {
			reportIntervalSec = val
		} else {
			fmt.Fprintf(os.Stderr, "Invalid REPORT_INTERVAL: %v\n", err)
			os.Exit(1)
		}
	}

	// Валидация
	if pollIntervalSec <= 0 || reportIntervalSec <= 0 {
		fmt.Fprintln(os.Stderr, "Poll and report intervals must be positive integers")
		os.Exit(1)
	}

	serverURL := "http://" + serverAddr
	pollInterval := time.Duration(pollIntervalSec) * time.Second
	reportInterval := time.Duration(reportIntervalSec) * time.Second

	pollsPerReport := reportIntervalSec / pollIntervalSec
	if pollsPerReport == 0 {
		pollsPerReport = 1
	}

	fmt.Printf("Starting agent: server=%s, poll=%v, report=%v\n",
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
