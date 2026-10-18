package repository

import (
	"context"

	"github.com/lebendig13/metrics/internal/config"
	models "github.com/lebendig13/metrics/internal/model"
)

type Storage interface {
	InsertOrUpdate(m models.Metrics) error
	Get(id string) (models.Metrics, bool)
	GetAllMetrics() map[string]string
	Snapshot() []models.Metrics
	Close() error
}

func NewStorage(ctx context.Context, configP *config.ServerConfig) (Storage, error) {
	if configP.DatabaseDSN == "" {
		return NewMemStorage(ctx, configP)
	}

	return NewDBStorage(configP.DatabaseDSN)
}
