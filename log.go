package app

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

func (c *App[T]) Logger() *zerolog.Logger {
	return &c.logger
}

func (c *App[T]) Log() *zerolog.Event {
	return c.logger.Info()
}

func (c *App[T]) Debug() *zerolog.Event {
	return c.logger.Debug()
}

func (c *App[T]) Error() *zerolog.Event {
	return c.logger.Error()
}

func (c *App[_]) makeLogger() {
	zerolog.TimeFieldFormat = time.RFC3339Nano

	c.logger = zerolog.New(os.Stdout)
	if c.cfg.Common.Debug {
		c.logger = c.logger.Level(zerolog.DebugLevel).Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "02.01.2006 15:04:05.000000",
		})
	} else {
		c.logger = c.logger.Level(zerolog.InfoLevel)
	}

	c.logger = c.logger.With().Timestamp().Logger()
	c.logger.Debug().Msg("logger: initialized")
	c.hasLogger = true
}
