package goapp

import (
	"errors"
	"os"
	"time"

	"github.com/rs/zerolog"
)

func (app *App[T]) Logger() *zerolog.Logger {
	return &app.f.logger
}

func (app *App[T]) Log() *zerolog.Event {
	return app.f.logger.Info()
}

func (app *App[T]) Warn() *zerolog.Event {
	return app.f.logger.Warn()
}

func (app *App[T]) Debug() *zerolog.Event {
	return app.f.logger.Debug()
}

func (app *App[T]) Error(errs ...error) *zerolog.Event {
	if len(errs) == 0 {
		return app.f.logger.Error()
	} else {
		return app.f.logger.Error().Err(errors.Join(errs...))
	}
}

func (app *App[_]) makeLogger() {
	zerolog.TimeFieldFormat = time.RFC3339Nano

	app.f.logger = zerolog.New(os.Stdout)
	if app.f.cfg.Common.Debug {
		app.f.logger = app.f.logger.Level(zerolog.DebugLevel).Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "02.01.2006 15:04:05.000000",
		})
	} else {
		app.f.logger = app.f.logger.Level(zerolog.InfoLevel)
	}

	app.f.logger = app.f.logger.With().Timestamp().Logger()
	app.f.logger.Debug().Msg("logger: initialized")
	app.f.hasLogger = true
}
