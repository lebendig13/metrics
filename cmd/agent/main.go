package main

import (
	"fmt"
	"log"
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
	log.Println("Server address: ", baseURL, "; pollInterval = ", pollInterval, "; reportInterval = ", reportInterval)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	rng := rand.New(rand.NewSource((time.Now().UnixNano())))
	agent := NewAgent()
	log.Println("Agent started")

	var timeSinceLastReport time.Duration
	for {
		time.Sleep(pollInterval)

		currentMetrics := agent.GetMetrics(rng)
		log.Println("Got all metrics. PollCount = ", agent.pollCount)

		timeSinceLastReport += pollInterval
		if timeSinceLastReport >= reportInterval {
			log.Println("Start sending metrics to server")
			handler.SendMetrics(client, currentMetrics, baseURL)

			timeSinceLastReport = 0
		}
	}
}
