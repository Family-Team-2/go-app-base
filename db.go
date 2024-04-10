package goapp

import (
	"context"
	"fmt"

	"github.com/Family-Team-2/go-app-base/database"
)

func (app *App[_]) DBC(ctx context.Context) *database.DB {
	return &database.DB{
		DB: app.f.db.WithContext(ctx),
	}
}

func (app *App[_]) DB() *database.DB {
	return app.DBC(app)
}

func (app *App[_]) UseSQLite() {
	app.SetDB(database.DB_SQLITE)
}

func (app *App[_]) UsePostgreSQL() {
	app.SetDB(database.DB_POSTGRES)
}

func (app *App[_]) UseCockroachDB() {
	app.SetDB(database.DB_CRDB)
}

func (app *App[_]) initDB() error {
	app, done := app.Perf("initDB")
	defer done()

	if !app.f.dbKind.HasDB() {
		app.f.db = nil
		return nil
	}

	db, err := database.NewDB(app.Context, &app.f.logger, app.f.dbKind, app.f.cfg.Common.DatabaseURL, app.f.cfg.Common.TraceSQL)
	if err != nil {
		return fmt.Errorf("creating db instance: %w", err)
	}

	version, err := db.GetVersion()
	if err != nil {
		app.f.logger.Error().Err(err).Msg("db: failed to get version")
	} else {
		app.f.logger.Debug().Str("version", version).Msg("db: running version")
	}

	err = db.SetMaxConnections(10)
	if err != nil {
		return fmt.Errorf("setting db max connections: %w", err)
	}

	app.f.db = db
	return nil
}

func (app *App[_]) shutdownDB() {
	app, done := app.Perf("shutdownDB")
	defer done()

	if app.f.db != nil {
		err := app.f.db.Shutdown()
		if err != nil {
			app.f.logger.Err(err).Msg("error while shutting down db")
		}
	}
}
