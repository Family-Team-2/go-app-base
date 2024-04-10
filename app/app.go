package app

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Family-Team-2/go-app-base/appctx"
	"github.com/Family-Team-2/go-app-base/database"
	"github.com/goccy/go-yaml"
	"github.com/rs/zerolog"
)

type CommonConfig struct {
	ConfigFile  string `yaml:"-"`
	Debug       bool   `yaml:"debug"`
	DatabaseURL string `yaml:"database_url"`
	TraceSQL    bool   `yaml:"trace_sql"`
}

type App[T any] struct {
	Cfg struct {
		Common CommonConfig `yaml:",inline"`
		Custom T            `yaml:",inline"`
	}

	dbKind  database.DBKind
	title   string
	version string

	ctx       context.Context
	logger    zerolog.Logger
	hasLogger bool
}

func (c *App[_]) CommonConfig() *CommonConfig {
	return &c.Cfg.Common
}

func (c *App[T]) CustomConfig() *T {
	return &c.Cfg.Custom
}

func (c *App[_]) SetTitle(title string) {
	c.title = title
}

func (c *App[_]) SetVersion(version string) {
	c.version = version
}

func (c *App[_]) SetDB(kind database.DBKind) {
	c.dbKind = kind
}

func (c *App[_]) Run(callback func(ctx *appctx.AppCtx) error) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	err := c.runContext(ctx, callback)
	if err == nil {
		c.logger.Info().Msg("shutting down")
	} else {
		if c.hasLogger {
			c.logger.Err(err).Msg("shutting down")
		} else {
			fmt.Println(err.Error())
		}

		os.Exit(1)
	}
}

func (c *App[_]) runContext(ctx context.Context, callback func(ctx *appctx.AppCtx) error) error {
	title := c.title
	if title == "" {
		title = "App"
	}

	version := c.version
	if version == "" {
		version = "0.0.1"
	}

	c.ctx = ctx

	flag.BoolVar(&c.Cfg.Common.Debug, "d", false, "")
	flag.BoolVar(&c.Cfg.Common.Debug, "debug", false, "")
	flag.StringVar(&c.Cfg.Common.ConfigFile, "c", "config.yml", "")
	flag.StringVar(&c.Cfg.Common.ConfigFile, "config-file", "config.yml", "")

	flag.Usage = func() {
		fmt.Println(title + " v" + version + "\n" +
			"Usage:\n" +
			"\t-d, --debug: enable debug output\n" +
			"\t-c, --config-file: path to config file")
	}

	flag.Parse()

	err := c.loadConfig()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	c.makeLogger()

	db, err := c.initDB()
	if err != nil {
		return fmt.Errorf("initializing DB: %w", err)
	}
	defer func() {
		if db != nil {
			err := db.Shutdown()
			if err != nil {
				c.logger.Err(err).Msg("error while shutting down db")
			}
		}
	}()

	ac := appctx.NewAppCtx(ctx, &c.logger, db)
	ac.Log().Str("title", c.title).Str("version", c.version).Msg("app: running")

	return callback(&ac)
}

func (c *App[_]) loadConfig() error {
	f, err := os.Open(c.Cfg.Common.ConfigFile)
	if err != nil {
		return fmt.Errorf("opening config file \"%v\": %w", c.Cfg.Common.ConfigFile, err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&c.Cfg)
	if err != nil {
		return fmt.Errorf("decoding YAML: %w", err)
	}

	return nil
}

func (c *App[_]) makeLogger() {
	zerolog.TimeFieldFormat = time.RFC3339Nano

	c.logger = zerolog.New(os.Stdout)
	if c.Cfg.Common.Debug {
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

func (c *App[_]) initDB() (*database.DB, error) {
	if !c.dbKind.HasDB() {
		return nil, nil
	}

	db, err := database.NewDB(c.ctx, &c.logger, c.dbKind, c.Cfg.Common.DatabaseURL, c.Cfg.Common.TraceSQL)
	if err != nil {
		return nil, fmt.Errorf("creating db instance: %w", err)
	}

	version, err := db.GetVersion()
	if err != nil {
		c.logger.Error().Err(err).Msg("db: failed to get version")
	} else {
		c.logger.Debug().Str("version", version).Msg("db: running version")
	}

	err = db.SetMaxConnections(10)
	if err != nil {
		return nil, fmt.Errorf("setting db max connections: %w", err)
	}

	return db, nil
}
