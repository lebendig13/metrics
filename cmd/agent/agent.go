package main

import (
	"math/rand"
	"runtime"

	models "github.com/lebendig13/metrics/internal/model"
)

type Agent struct {
	pollCount int64
}

func NewAgent() *Agent {
	return &Agent{}
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
	newPollCount := &models.Metrics{
		ID:    "PollCount",
		MType: models.Counter,
	}
	newPollCount.Delta = &a.pollCount

	result = append(result, newPollCount)
	return result
}
