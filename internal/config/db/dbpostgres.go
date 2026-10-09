package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type SQLDB struct {
	db *sql.DB
}

func InitDBConnection(databaseDSN string) (*SQLDB, error) {
	db, err := sql.Open("pgx", databaseDSN)
	if err != nil {
		return nil, fmt.Errorf("cannot open database: %w", err)
	}

	sqldb := &SQLDB{db: db}

	if err := sqldb.CheckDBConnection(); err != nil {
		db.Close()
		return nil, err
	}

	return sqldb, nil
}

func (d *SQLDB) CheckDBConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if d.db == nil {
		return fmt.Errorf("No database")
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
