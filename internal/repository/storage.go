package repository

import (
	"context"
	"strings"

	"github.com/lebendig13/metrics/internal/config"
	models "github.com/lebendig13/metrics/internal/model"
)

type Storage interface {
	InsertOrUpdate(ctx context.Context, m models.Metrics) error
	UpdateMetricsBatch(ctx context.Context, metrics []models.Metrics) error
	Get(ctx context.Context, id string) (models.Metrics, bool)
	GetAllMetrics(ctx context.Context) map[string]string
	Snapshot(ctx context.Context) []models.Metrics
	Close(ctx context.Context) error
	CheckDBConnection() error
}

func NewStorage(ctx context.Context, configP *config.ServerConfig) (Storage, error) {
	if configP.DatabaseDSN != "" && strings.HasPrefix(configP.DatabaseDSN, "postgres://") {
		return NewDBStorage(configP.DatabaseDSN)
	}

	return NewMemStorage(ctx, configP)
}
