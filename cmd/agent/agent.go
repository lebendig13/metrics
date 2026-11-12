package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"

	handler "github.com/lebendig13/metrics/internal/agent"
	config "github.com/lebendig13/metrics/internal/config"
	models "github.com/lebendig13/metrics/internal/model"
)

const sendTimeout = 15 * time.Second

type Intervals struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
}

type Agent struct {
	baseURL   string
	intervals Intervals
	pollCount int64
	hashKey   string
	rateLimit int
}

type sendJob struct {
	metrics   []*models.Metrics
	pollCount int64
}

func NewAgentDefault() *Agent {
	return &Agent{
		baseURL: "http://localhost:8080/updates/",
		intervals: Intervals{
			PollInterval:   time.Duration(2) * time.Second,
			ReportInterval: time.Duration(10) * time.Second,
		},
		pollCount: 0,
		hashKey:   "",
		rateLimit: 1,
	}
}

func NewAgent(cnf *config.AgentConfig) *Agent {
	return &Agent{
		baseURL: fmt.Sprintf("http://%s/updates/", cnf.ServerAddress),
		intervals: Intervals{
			PollInterval:   time.Duration(cnf.Intervals.PollInterval) * time.Second,
			ReportInterval: time.Duration(cnf.Intervals.ReportInterval) * time.Second,
		},
		pollCount: 0,
		hashKey:   cnf.HashKey,
		rateLimit: cnf.RateLimit,
	}
}

func (a *Agent) Process(ctx context.Context) {
	log.Println("Agent started")
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	sendJobs := make(chan sendJob, a.rateLimit)
	results := make(chan int64, a.rateLimit)

	var workersWg sync.WaitGroup
	for i := 1; i <= a.rateLimit; i++ {
		workersWg.Add(1)
		go func(id int) {
			defer workersWg.Done()
			a.sendWorker(ctx, id, client, sendJobs, results)
		}(i)
	}

	systemMetrics := make(chan []*models.Metrics, 1)

	var collectorsWg sync.WaitGroup
	collectorsWg.Add(2)
	go func() {
		defer collectorsWg.Done()
		a.collectSystem(ctx, systemMetrics)
	}()
	go func() {
		defer collectorsWg.Done()
		a.collect(ctx, sendJobs, results, systemMetrics)
	}()

	collectorsWg.Wait()
	close(sendJobs)
	workersWg.Wait()
	log.Println("Agent stopped")
}

func (a *Agent) collect(ctx context.Context, jobs chan<- sendJob, result <-chan int64, systemMetrics <-chan []*models.Metrics) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	pollTicker := time.NewTicker(a.intervals.PollInterval)
	defer pollTicker.Stop()
	reportTicker := time.NewTicker(a.intervals.ReportInterval)
	defer reportTicker.Stop()

	var currentMetrics []*models.Metrics
	var currentSystemMetrics []*models.Metrics
	for {
		select {
		case <-ctx.Done():
			return

		case n := <-result:
			a.pollCount += n

		case currentSystemMetrics = <-systemMetrics:

		case <-pollTicker.C:
			currentMetrics = a.GetMetrics(rng)
			log.Println("Got all metrics. PollCount = ", a.pollCount)

		case <-reportTicker.C:
			if currentMetrics == nil {
				continue
			}
			job := a.newSendJob(currentMetrics, currentSystemMetrics)
			select {
			case jobs <- job:
				a.pollCount = 0
				log.Println("Start sending metrics to server")
			default:
				log.Println("All workers are busy, skip sending metrics")
			}
		}
	}
}

func (a *Agent) newSendJob(snapshot []*models.Metrics, systemSnapshot []*models.Metrics) sendJob {
	pollCount := a.pollCount
	metrics := make([]*models.Metrics, 0, len(snapshot)+len(systemSnapshot))
	for _, m := range snapshot {
		if m.ID != "PollCount" {
			metrics = append(metrics, m)
		}
	}
	metrics = append(metrics, systemSnapshot...)
	metrics = append(metrics, &models.Metrics{
		ID:    "PollCount",
		MType: models.Counter,
		Delta: &pollCount,
	})
	return sendJob{metrics: metrics, pollCount: pollCount}
}

func (a *Agent) sendWorker(ctx context.Context, id int, client *http.Client, jobs <-chan sendJob, result chan<- int64) {
	for job := range jobs {
		sendCtx, cancel := context.WithTimeout(ctx, sendTimeout)
		err := handler.SendMetricsBatch(sendCtx, client, job.metrics, a.baseURL, a.hashKey)
		cancel()

		if err == nil {
			log.Printf("Worker %d: metrics batch sent", id)
			continue
		}

		log.Printf("Worker %d: failed to send metrics batch: %v", id, err)
		select {
		case result <- job.pollCount:
		case <-ctx.Done():
		}
	}
}

func (a *Agent) collectSystem(ctx context.Context, out chan []*models.Metrics) {
	ticker := time.NewTicker(a.intervals.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			metrics, err := GetSystemMetrics(ctx)
			if err != nil {
				log.Println("Cannot get system metrics: ", err)
				continue
			}
			select {
			case <-out:
			default:
			}
			out <- metrics
		}
	}
}

func GetSystemMetrics(ctx context.Context) ([]*models.Metrics, error) {
	vm, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot get memory stats: %w", err)
	}
	cpuPercents, err := cpu.PercentWithContext(ctx, 0, true)
	if err != nil {
		return nil, fmt.Errorf("cannot get CPU utilization: %w", err)
	}

	result := make([]*models.Metrics, 0, 2+len(cpuPercents))
	newGauge := func(id string, value float64) *models.Metrics {
		return &models.Metrics{ID: id, MType: models.Gauge, Value: &value}
	}
	result = append(result,
		newGauge("TotalMemory", float64(vm.Total)),
		newGauge("FreeMemory", float64(vm.Free)),
	)
	for i, p := range cpuPercents {
		result = append(result, newGauge("CPUutilization"+strconv.Itoa(i+1), p))
	}
	return result, nil
}

func (a *Agent) GetMetrics(rng *rand.Rand) []*models.Metrics {
	var rms runtime.MemStats
	runtime.ReadMemStats(&rms)

	result := make([]*models.Metrics, 0)

	for name, value := range map[string]float64{
		"Alloc":         float64(rms.Alloc),
		"BuckHashSys":   float64(rms.BuckHashSys),
		"Frees":         float64(rms.Frees),
		"GCCPUFraction": float64(rms.GCCPUFraction),
		"GCSys":         float64(rms.GCSys),
		"HeapAlloc":     float64(rms.HeapAlloc),
		"HeapIdle":      float64(rms.HeapIdle),
		"HeapInuse":     float64(rms.HeapInuse),
		"HeapObjects":   float64(rms.HeapObjects),
		"HeapReleased":  float64(rms.HeapReleased),
		"HeapSys":       float64(rms.HeapSys),
		"LastGC":        float64(rms.LastGC),
		"Lookups":       float64(rms.Lookups),
		"MCacheInuse":   float64(rms.MCacheInuse),
		"MCacheSys":     float64(rms.MCacheSys),
		"MSpanInuse":    float64(rms.MSpanInuse),
		"MSpanSys":      float64(rms.MSpanSys),
		"Mallocs":       float64(rms.Mallocs),
		"NextGC":        float64(rms.NextGC),
		"NumForcedGC":   float64(rms.NumForcedGC),
		"NumGC":         float64(rms.NumGC),
		"OtherSys":      float64(rms.OtherSys),
		"PauseTotalNs":  float64(rms.PauseTotalNs),
		"StackInuse":    float64(rms.StackInuse),
		"StackSys":      float64(rms.StackSys),
		"Sys":           float64(rms.Sys),
		"TotalAlloc":    float64(rms.TotalAlloc),
		"RandomValue":   rng.Float64(),
	} {
		metric := &models.Metrics{
			ID:    name,
			MType: models.Gauge,
		}
		metric.Value = &value
		result = append(result, metric)
	}

	a.pollCount++
	pollCount := a.pollCount
	newPollCount := &models.Metrics{
		ID:    "PollCount",
		MType: models.Counter,
		Delta: &pollCount,
	}

	result = append(result, newPollCount)
	return result
}
