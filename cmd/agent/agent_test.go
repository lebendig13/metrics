package main

import (
	"context"
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	models "github.com/lebendig13/metrics/internal/model"
)

func TestGetSystemMetrics(t *testing.T) {
	metrics, err := GetSystemMetrics(context.Background())
	require.NoError(t, err)

	byID := make(map[string]*models.Metrics)
	cpuCount := 0
	for _, m := range metrics {
		assert.Equal(t, models.Gauge, m.MType)
		require.NotNil(t, m.Value)
		byID[m.ID] = m
		if strings.HasPrefix(m.ID, "CPUutilization") {
			cpuCount++
		}
	}

	require.Contains(t, byID, "TotalMemory")
	require.Contains(t, byID, "FreeMemory")
	assert.Greater(t, *byID["TotalMemory"].Value, 0.0)
	assert.LessOrEqual(t, *byID["FreeMemory"].Value, *byID["TotalMemory"].Value)

	assert.GreaterOrEqual(t, cpuCount, 1)
	assert.Contains(t, byID, "CPUutilization1")
}

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
