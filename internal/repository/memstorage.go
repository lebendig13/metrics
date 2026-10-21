package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/lebendig13/metrics/internal/config"
	"github.com/lebendig13/metrics/internal/logger"
	models "github.com/lebendig13/metrics/internal/model"
)

type MemStorage struct {
	metrics         map[string]models.Metrics
	fileStoragePath string
}

func NewMemStorage(ctx context.Context, configP *config.ServerConfig) (*MemStorage, error) {
	mst := MemStorage{
		metrics:         make(map[string]models.Metrics),
		fileStoragePath: configP.FileStoragePath,
	}

	ProcessBackup(ctx, &mst, configP)

	return &mst, nil
}

func (ms *MemStorage) InsertOrUpdate(m models.Metrics) error {
	if m.ID == "" {
		return errors.New("empty metric ID")
	}

	if ms.metrics == nil {
		ms.metrics = make(map[string]models.Metrics)
	}

	switch m.MType {
	case models.Counter:
		if m.Delta == nil {
			return errors.New("counter: nil value, nothing to change")
		}
		currentMetric, exists := ms.metrics[m.ID]
		if exists && currentMetric.Delta != nil {
			newMetric := *currentMetric.Delta + *m.Delta
			m.Delta = &newMetric
		}
		ms.metrics[m.ID] = m
	case models.Gauge:
		if m.Value == nil {
			return errors.New("gauge: nil value, nothing to change")
		}
		ms.metrics[m.ID] = m
	default:
		return errors.New("unknown metric type")
	}

	return nil
}

func (ms *MemStorage) UpdateMetricsBatch(metrics []models.Metrics) error {
	for _, m := range metrics {
		if m.ID == "" {
			return errors.New("empty metric ID")
		}

		if ms.metrics == nil {
			ms.metrics = make(map[string]models.Metrics)
		}

		switch m.MType {
		case models.Counter:
			if m.Delta == nil {
				return errors.New("counter: nil value, nothing to change")
			}
			currentMetric, exists := ms.metrics[m.ID]
			if exists && currentMetric.Delta != nil {
				newMetric := *currentMetric.Delta + *m.Delta
				m.Delta = &newMetric
			}
			ms.metrics[m.ID] = m
		case models.Gauge:
			if m.Value == nil {
				return errors.New("gauge: nil value, nothing to change")
			}
			ms.metrics[m.ID] = m
		default:
			return errors.New("unknown metric type")
		}
	}
	return nil
}

func (ms *MemStorage) Get(id string) (models.Metrics, bool) {
	currentMetric, exists := ms.metrics[id]
	return currentMetric, exists
}

func (ms *MemStorage) GetAllMetrics() map[string]string {
	result := make(map[string]string)
	for _, m := range ms.metrics {
		switch m.MType {
		case models.Counter:
			if m.Delta == nil {
				continue
			}
			result[m.ID] = strconv.FormatInt(*m.Delta, 10)
		case models.Gauge:
			if m.Value == nil {
				continue
			}
			result[m.ID] = strconv.FormatFloat(*m.Value, 'f', 10, 64)
		}
	}
	return result
}

func (ms *MemStorage) Snapshot() []models.Metrics {
	var result []models.Metrics
	for _, m := range ms.metrics {
		result = append(result, m)
	}
	return result
}

func (ms *MemStorage) Close() error {
	if ms.fileStoragePath == "" {
		return nil
	}

	logger.Log.Info("Saving current metrics before shutdown...")
	currentMetrics := ms.Snapshot()

	err := SaveMetrics(currentMetrics, ms.fileStoragePath)
	if err != nil {
		return fmt.Errorf("cannot save current metrics: %w", err)
	}

	logger.Log.Info("Current metrics saved successfully")
	return nil
}
