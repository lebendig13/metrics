package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	db "github.com/lebendig13/metrics/internal/config/db"
	models "github.com/lebendig13/metrics/internal/model"
)

type Pingable interface {
	CheckDBConnection() error
}

type DBStorage struct {
	sqldb *db.SQLDB
}

func NewDBStorage(dbDSN string) (*DBStorage, error) {

	sdb, err := db.InitDBConnection(dbDSN)
	if err != nil {
		return nil, err
	}

	dbst := DBStorage{
		sqldb: sdb,
	}

	return &dbst, nil
}

func (dbs *DBStorage) InsertOrUpdate(m models.Metrics) error {
	if m.ID == "" {
		return errors.New("empty metric ID")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		INSERT INTO metrics (id, type, delta, value)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id, type) 
		DO UPDATE SET 
			delta = CASE WHEN metrics.type = 'counter' THEN metrics.delta + EXCLUDED.delta ELSE EXCLUDED.delta END,
			value = CASE WHEN metrics.type = 'gauge' THEN EXCLUDED.value ELSE EXCLUDED.value END;
	`
	_, err := dbs.sqldb.ExecContext(ctx, query, m.ID, m.MType, m.Delta, m.Value)
	if err != nil {
		return fmt.Errorf("failed to insert or update metric %s: %w", m.ID, err)
	}

	return nil
}

func (dbs *DBStorage) Get(id string) (models.Metrics, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT id, type, delta, value FROM metrics WHERE id = $1;`

	var m models.Metrics
	var delta sql.NullInt64
	var value sql.NullFloat64

	err := dbs.sqldb.QueryRowContext(ctx, query, id).Scan(&m.ID, &m.MType, &delta, &value)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Metrics{}, false
		}
		return models.Metrics{}, false
	}

	if delta.Valid {
		m.Delta = &delta.Int64
	}
	if value.Valid {
		m.Value = &value.Float64
	}

	return m, true
}

func (dbs *DBStorage) GetAllMetrics() map[string]string {
	result := make(map[string]string)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT id, type, delta, value FROM metrics;`
	rows, err := dbs.sqldb.QueryContext(ctx, query)
	if err != nil {
		return result
	}
	defer rows.Close()

	for rows.Next() {
		var id, mType string
		var delta sql.NullInt64
		var value sql.NullFloat64

		if err := rows.Scan(&id, &mType, &delta, &value); err != nil {
			continue
		}

		if mType == "counter" && delta.Valid {
			result[id] = fmt.Sprintf("%d", delta.Int64)
		} else if mType == "gauge" && value.Valid {
			result[id] = fmt.Sprintf("%g", value.Float64)
		}
	}

	if err := rows.Err(); err != nil {
		return result
	}

	return result
}

func (dbs *DBStorage) Snapshot() []models.Metrics {
	var result []models.Metrics

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT id, type, delta, value FROM metrics;`
	rows, err := dbs.sqldb.QueryContext(ctx, query)
	if err != nil {
		return result
	}
	defer rows.Close()

	for rows.Next() {
		var m models.Metrics
		var delta sql.NullInt64
		var value sql.NullFloat64

		if err := rows.Scan(&m.ID, &m.MType, &delta, &value); err != nil {
			continue
		}

		if delta.Valid {
			m.Delta = &delta.Int64
		}
		if value.Valid {
			m.Value = &value.Float64
		}

		result = append(result, m)
	}

	if err := rows.Err(); err != nil {
		return result
	}

	return result
}

func (dbs *DBStorage) Close() error {
	if dbs.sqldb != nil {
		return dbs.sqldb.Close()
	}

	return nil
}

func (dbs *DBStorage) CheckDBConnection() error {
	return dbs.sqldb.CheckDBConnection()
}
