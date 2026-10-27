package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	db "github.com/lebendig13/metrics/internal/config/db"
	models "github.com/lebendig13/metrics/internal/model"
)

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

func (dbs *DBStorage) InsertOrUpdate(ctx context.Context, m models.Metrics) error {
	if m.ID == "" {
		return errors.New("empty metric ID")
	}

	query := `
		INSERT INTO metrics (id, type, delta, value)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id, type) 
		DO UPDATE SET 
			delta = CASE WHEN metrics.type = 'counter' THEN metrics.delta + EXCLUDED.delta ELSE metrics.delta END,
			value = EXCLUDED.value;
	`
	_, err := dbs.sqldb.ExecContext(ctx, query, m.ID, m.MType, m.Delta, m.Value)
	if err != nil {
		return fmt.Errorf("failed to insert or update metric %s: %w", m.ID, err)
	}

	return nil

}

func (dbs *DBStorage) UpdateMetricsBatch(ctx context.Context, metrics []models.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}

	attemptIntervals := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}
	var lastErr error

	query := `
		INSERT INTO metrics (id, type, delta, value)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id, type) 
		DO UPDATE SET 
			delta = CASE WHEN metrics.type = 'counter' THEN metrics.delta + EXCLUDED.delta ELSE metrics.delta END,
			value = EXCLUDED.value;
	`

	for i := 0; i <= len(attemptIntervals); i++ {
		err := func() error {
			tx, err := dbs.sqldb.BeginTx(ctx)
			if err != nil {
				return fmt.Errorf("cannot begin transaction: %w", err)
			}
			defer func() {
				_ = tx.Rollback()
			}()

			stmt, err := tx.PrepareContext(ctx, query)
			if err != nil {
				return err
			}
			defer stmt.Close()

			for _, m := range metrics {
				if m.ID == "" {
					return errors.New("empty metric ID")
				}
				if _, err := stmt.ExecContext(ctx, m.ID, m.MType, m.Delta, m.Value); err != nil {
					return err
				}
			}

			return tx.Commit()
		}()

		if err == nil {
			return nil
		}

		lastErr = err

		if !isPostgresRetriable(lastErr) || i == len(attemptIntervals) {
			break
		}

		select {
		case <-time.After(attemptIntervals[i]):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return fmt.Errorf("failed to update batch after retries: %w", lastErr)
}

func (dbs *DBStorage) Get(ctx context.Context, id string) (models.Metrics, bool) {
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

func (dbs *DBStorage) GetAllMetrics(ctx context.Context) map[string]string {
	result := make(map[string]string)

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

func (dbs *DBStorage) Snapshot(ctx context.Context) []models.Metrics {
	var result []models.Metrics

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

func (dbs *DBStorage) Close(ctx context.Context) error {
	if dbs.sqldb != nil {
		return dbs.sqldb.Close()
	}

	return nil
}

func (dbs *DBStorage) CheckDBConnection() error {
	return dbs.sqldb.CheckDBConnection()
}

func isPostgresRetriable(err error) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return strings.HasPrefix(pgErr.Code, "08")
	}

	errMsg := err.Error()
	return strings.Contains(errMsg, "connection refused") ||
		strings.Contains(errMsg, "broken pipe") ||
		strings.Contains(errMsg, "connection busy")
}
