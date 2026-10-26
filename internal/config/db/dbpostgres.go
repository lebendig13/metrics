package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DBConnection interface {
	CheckDBConnection() error
}

type SQLDB struct {
	db *sql.DB
}

func InitDBConnection(databaseDSN string) (*SQLDB, error) {
	if databaseDSN == "" {
		return nil, fmt.Errorf("database DSN is empty")
	}

	db, err := sql.Open("pgx", databaseDSN)
	if err != nil {
		return nil, fmt.Errorf("cannot open database: %w", err)
	}

	// ping БД для проверки, что все ок
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("database ping failed on init: %w", err)
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

func (d *SQLDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if d.db == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}
	return d.db.ExecContext(ctx, query, args...)
}

func (d *SQLDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if d.db == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}
	return d.db.QueryContext(ctx, query, args...)
}

func (d *SQLDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	if d.db == nil {
		return nil
	}
	return d.db.QueryRowContext(ctx, query, args...)
}

func (d *SQLDB) BeginTx(ctx context.Context) (*sql.Tx, error) {
	if d.db == nil {
		return nil, fmt.Errorf("nil db")
	}
	return d.db.BeginTx(ctx, nil)
}
