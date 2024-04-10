package main

import (
	"context"
	"fmt"

	"github.com/Family-Team-2/go-app-base/database"
)

func (c *App[_]) initDB() error {
	if !c.dbKind.HasDB() {
		c.db = nil
		return nil
	}

	db, err := database.NewDB(c.Context, &c.logger, c.dbKind, c.cfg.Common.DatabaseURL, c.cfg.Common.TraceSQL)
	if err != nil {
		return fmt.Errorf("creating db instance: %w", err)
	}

	version, err := db.GetVersion()
	if err != nil {
		c.logger.Error().Err(err).Msg("db: failed to get version")
	} else {
		c.logger.Debug().Str("version", version).Msg("db: running version")
	}

	err = db.SetMaxConnections(10)
	if err != nil {
		return fmt.Errorf("setting db max connections: %w", err)
	}

	c.db = db
	return nil
}

func (c *App[_]) shutdownDB() {
	if c.db != nil {
		err := c.db.Shutdown()
		if err != nil {
			c.logger.Err(err).Msg("error while shutting down db")
		}
	}
}

func (c *App[_]) DBC(ctx context.Context) *database.DB {
	return &database.DB{
		DB: c.db.WithContext(ctx),
	}
}

func (c *App[_]) DB() *database.DB {
	return c.DBC(c)
}

func (c *App[_]) UseSQLite() {
	c.SetDB(database.DB_SQLITE)
}

func (c *App[_]) UsePostgreSQL() {
	c.SetDB(database.DB_POSTGRES)
}

func (c *App[_]) UseCockroachDB() {
	c.SetDB(database.DB_CRDB)
}
