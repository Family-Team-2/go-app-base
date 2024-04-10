package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach-go/v2/crdb/crdbgorm"
	"github.com/glebarez/sqlite"
	"github.com/rs/zerolog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DBKind int

const (
	DB_NONE DBKind = iota
	DB_SQLITE
	DB_POSTGRES
	DB_CRDB
)

func (k DBKind) HasDB() bool {
	return k != DB_NONE
}

type DB struct {
	*gorm.DB
	kind DBKind
}

func NewDB(ctx context.Context, logger *zerolog.Logger, kind DBKind, databaseURL string, trace bool) (*DB, error) {
	if kind == DB_NONE {
		return nil, nil
	}

	db := DB{
		kind: kind,
	}
	var dialector gorm.Dialector
	switch db.kind {
	case DB_SQLITE:
		dialector = sqlite.Open(databaseURL)
	case DB_CRDB, DB_POSTGRES:
		dialector = postgres.Open(databaseURL)
	}

	var err error
	db.DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: newDBLogger(logger, trace),
	})
	if err != nil {
		return nil, fmt.Errorf("opening db: %w", err)
	}

	sqlDB, err := db.sqlDB()
	if err != nil {
		return nil, fmt.Errorf("getting SQL DB: %w", err)
	}

	sqlDB.SetConnMaxLifetime(time.Minute * 5)

	err = db.testFeatures(ctx)
	if err != nil {
		return nil, fmt.Errorf("testing db features: %w", err)
	}

	return &db, nil
}

func (db *DB) Kind() DBKind {
	return db.kind
}

func (db *DB) SetMaxConnections(maxConnections int) error {
	sqlDB, err := db.sqlDB()
	if err != nil {
		return fmt.Errorf("getting SQL DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(min(4, maxConnections))
	sqlDB.SetMaxOpenConns(maxConnections)
	return nil
}

func (db *DB) Shutdown() error {
	sqlDB, err := db.sqlDB()
	if err != nil {
		return fmt.Errorf("getting SQL DB: %w", err)
	}

	err = sqlDB.Close()
	if err != nil {
		return fmt.Errorf("closing DB: %w", err)
	}

	db.DB = nil
	return nil
}

func (db *DB) sqlDB() (*sql.DB, error) {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return nil, err
	}

	return sqlDB, nil
}

func (db *DB) Transaction(ctx context.Context, callback func(tx *gorm.DB) error) error {
	if db.kind == DB_CRDB {
		return crdbgorm.ExecuteTx(ctx, db.DB, nil, callback)
	} else {
		return db.DB.Transaction(callback)
	}
}
