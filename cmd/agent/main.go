package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/lebendig13/metrics/internal/handler"
)

var (
	baseURL        string
	pollInterval   time.Duration
	reportInterval time.Duration
)

func main() {
	parseFlags()
	baseURL = fmt.Sprintf("http://%s/update/", flagServerAddr)
	pollInterval = time.Duration(flagPollInterval) * time.Second
	reportInterval = time.Duration(flagReportInterval) * time.Second
	fmt.Println("Server address: ", baseURL, "; pollInterval = ", pollInterval, "; reportInterval = ", reportInterval)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	rng := rand.New(rand.NewSource((time.Now().UnixNano())))
	agent := NewAgent()
	fmt.Println("Agent started")

	var timeSinceLastReport time.Duration
	for {
		time.Sleep(pollInterval)

		currentMetrics := agent.GetMetrics(rng)
		fmt.Println("Got all metrics. PollCount = ", agent.pollCount)

		timeSinceLastReport += pollInterval
		if timeSinceLastReport >= reportInterval {
			fmt.Println("Start sending metrics to server")
			handler.SendMetrics(client, currentMetrics, baseURL)

			timeSinceLastReport = 0
		}
	}
}
