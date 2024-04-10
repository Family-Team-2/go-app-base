package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

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
	context.Context

	cfg struct {
		Common CommonConfig `yaml:",inline"`
		Custom T            `yaml:",inline"`
	}

	db      *database.DB
	dbKind  database.DBKind
	title   string
	version string

	logger    zerolog.Logger
	hasLogger bool
}

func (c *App[_]) CommonConfig() *CommonConfig {
	return &c.cfg.Common
}

func (c *App[T]) Config() *T {
	return &c.cfg.Custom
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

func (c *App[T]) Run(callback func(ctx *App[T]) error) {
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

func (c *App[T]) runContext(ctx context.Context, callback func(ctx *App[T]) error) error {
	title := c.title
	if title == "" {
		title = "App"
	}

	version := c.version
	if version == "" {
		version = "0.0.1"
	}

	c.Context = ctx

	flag.BoolVar(&c.cfg.Common.Debug, "d", false, "")
	flag.BoolVar(&c.cfg.Common.Debug, "debug", false, "")
	flag.StringVar(&c.cfg.Common.ConfigFile, "c", "config.yml", "")
	flag.StringVar(&c.cfg.Common.ConfigFile, "config-file", "config.yml", "")

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

	err = c.initDB()
	if err != nil {
		return fmt.Errorf("initializing DB: %w", err)
	}
	defer c.shutdownDB()

	c.Log().Str("title", c.title).Str("version", c.version).Msg("app: running")
	return callback(c)
}

func (c *App[_]) loadConfig() error {
	f, err := os.Open(c.cfg.Common.ConfigFile)
	if err != nil {
		return fmt.Errorf("opening config file \"%v\": %w", c.cfg.Common.ConfigFile, err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&c.cfg)
	if err != nil {
		return fmt.Errorf("decoding YAML: %w", err)
	}

	return nil
}

func (c *App[T]) clone() *App[T] {
	var newApp = *c
	return &newApp
}
