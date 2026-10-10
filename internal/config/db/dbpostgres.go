package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/lebendig13/metrics/internal/logger"
)

type SQLDB struct {
	db *sql.DB
}

func InitDBConnection(databaseDSN string) (*SQLDB, error) {
	if databaseDSN == "" {
		if logger.Log != nil {
			logger.Log.Info("Database DSN is empty")
		} else {
			fmt.Println("Database DSN is empty")
		}

		return nil, nil
	}

	db, err := sql.Open("pgx", databaseDSN)
	if err != nil {
		return nil, fmt.Errorf("cannot open database: %w", err)
	}

	sqldb := &SQLDB{db: db}
	return sqldb, nil
}

func (d *SQLDB) CheckDBConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if d.db == nil {
		return fmt.Errorf("no database")
	}

	if err := d.db.PingContext(ctx); err != nil {
		return fmt.Errorf("cannot connect to database: %w", err)
	}
	return nil
}

func (d *SQLDB) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}
