package models

import (
	"errors"
	"strconv"
)

const (
	Counter = "counter"
	Gauge   = "gauge"
)

// NOTE: Не усложняем пример, вводя иерархическую вложенность структур.
// Органичиваясь плоской моделью.
// Delta и Value объявлены через указатели,
// что бы отличать значение "0", от не заданного значения
// и соответственно не кодировать в структуру.
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}

type MemStorage struct {
	metrics map[string]Metrics
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		metrics: make(map[string]Metrics),
	}
}

func (ms *MemStorage) InsertOrUpdate(m Metrics) error {
	switch m.MType {
	case Counter:
		if m.Delta == nil {
			return errors.New("counter: nil value, nothing to change")
		}
		currentMetric, exists := ms.metrics[m.ID]
		if exists && currentMetric.Delta != nil {
			newMetric := *currentMetric.Delta + *m.Delta
			m.Delta = &newMetric
		}
		ms.metrics[m.ID] = m
	case Gauge:
		if m.Value == nil {
			return errors.New("gauge: nil value, nothing to change")
		}
		ms.metrics[m.ID] = m
	default:
		return errors.New("unknown metric type")
	}

	return nil
}

func (ms *MemStorage) Get(id string) (Metrics, bool) {
	currentMetric, exists := ms.metrics[id]
	return currentMetric, exists
}

func (ms *MemStorage) GetAllMetrics() map[string]string {
	result := make(map[string]string)
	for _, m := range ms.metrics {
		switch m.MType {
		case Counter:
			if m.Delta == nil {
				continue
			}
			result[m.ID] = strconv.FormatInt(*m.Delta, 10)
		case Gauge:
			if m.Value == nil {
				continue
			}
			result[m.ID] = strconv.FormatFloat(*m.Value, 'f', 10, 64)
		}
	}
	return result
}
