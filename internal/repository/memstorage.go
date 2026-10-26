package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/lebendig13/metrics/internal/config"
	"github.com/lebendig13/metrics/internal/logger"
	models "github.com/lebendig13/metrics/internal/model"
)

type MemStorage struct {
	mx              sync.RWMutex
	metrics         map[string]models.Metrics
	fileStoragePath string
	databaseDSN     string
}

func NewMemStorage(ctx context.Context, configP *config.ServerConfig) (*MemStorage, error) {
	mst := MemStorage{
		metrics:         make(map[string]models.Metrics),
		fileStoragePath: configP.FileStoragePath,
		databaseDSN:     configP.DatabaseDSN,
	}

	ProcessBackup(ctx, &mst, configP)

	return &mst, nil
}

func (ms *MemStorage) InsertOrUpdate(ctx context.Context, m models.Metrics) error {
	if m.ID == "" {
		return errors.New("empty metric ID")
	}

	ms.mx.Lock()
	defer ms.mx.Unlock()

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

func (ms *MemStorage) UpdateMetricsBatch(ctx context.Context, metrics []models.Metrics) error {
	ms.mx.Lock()
	defer ms.mx.Unlock()

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

func (ms *MemStorage) Get(ctx context.Context, id string) (models.Metrics, bool) {
	ms.mx.RLock()
	defer ms.mx.RUnlock()

	currentMetric, exists := ms.metrics[id]
	return currentMetric, exists
}

func (ms *MemStorage) GetAllMetrics(ctx context.Context) map[string]string {
	ms.mx.RLock()
	defer ms.mx.RUnlock()

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

func (ms *MemStorage) Snapshot(ctx context.Context) []models.Metrics {
	ms.mx.RLock()
	defer ms.mx.RUnlock()

	var result []models.Metrics
	for _, m := range ms.metrics {
		result = append(result, m)
	}
	return result
}

func (ms *MemStorage) Close(ctx context.Context) error {
	if ms.fileStoragePath == "" {
		return nil
	}

	logger.Log.Info("Saving current metrics before shutdown...")
	currentMetrics := ms.Snapshot(ctx)

	err := SaveMetrics(currentMetrics, ms.fileStoragePath)
	if err != nil {
		return fmt.Errorf("cannot save current metrics: %w", err)
	}

	logger.Log.Info("Current metrics saved successfully")
	return nil
}

func (ms *MemStorage) CheckDBConnection() error {
	if ms.databaseDSN != "" {
		return fmt.Errorf("database DSN was provided (%s), but server uses MemStorage", ms.databaseDSN)
	}
	return nil
}
