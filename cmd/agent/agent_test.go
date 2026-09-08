package main

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMetrics(t *testing.T) {
	// Начальное состояние
	agent := NewAgentDefault()
	assert.Equal(t, int64(0), agent.pollCount)

	// Первое получение метрик
	rng := rand.New(rand.NewSource(100))
	metrics := agent.GetMetrics(rng)

	// Счетчик метрик увеличился
	assert.Equal(t, int64(1), agent.pollCount)

	// Проверяем наличие всех метрик
	allMetricNames := []string{
		"Alloc",
		"BuckHashSys",
		"Frees",
		"GCCPUFraction",
		"GCSys",
		"HeapAlloc",
		"HeapIdle",
		"HeapInuse",
		"HeapObjects",
		"HeapReleased",
		"HeapSys",
		"LastGC",
		"Lookups",
		"MCacheInuse",
		"MCacheSys",
		"MSpanInuse",
		"MSpanSys",
		"Mallocs",
		"NextGC",
		"NumForcedGC",
		"NumGC",
		"OtherSys",
		"PauseTotalNs",
		"StackInuse",
		"StackSys",
		"Sys",
		"TotalAlloc",
		"RandomValue",
		"NonExistentMetric", // для проверки пропущенных метрик
	}
	var exists bool
	for _, metricName := range allMetricNames {
		exists = false
		for _, metric := range metrics {
			if metricName == metric.ID {
				exists = true

				// Проверка значения RandomValue
				if metricName == "RandomValue" {
					assert.Less(t, *metric.Value, 1.0)
					assert.Greater(t, *metric.Value, 0.0)
				}

				break
			}
		}
		assert.Equal(t, (metricName != "NonExistentMetric"), exists)
	}

	// Повторное получение метрик
	agent.GetMetrics(rng)

	// Счетчик метрик увеличился
	assert.Equal(t, int64(2), agent.pollCount)
}
