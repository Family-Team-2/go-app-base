package goapp

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
	TracePerf   bool   `yaml:"trace_perf"`
}

type appFields[T any] struct {
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

type App[T any] struct {
	context.Context

	f *appFields[T]
}

func NewApp[T any](title, version string) *App[T] {
	return &App[T]{
		f: &appFields[T]{
			title:   title,
			version: version,
		},
	}
}

func (app *App[_]) CommonConfig() *CommonConfig {
	app.ensureFieldsSet()
	return &app.f.cfg.Common
}

func (app *App[T]) Config() *T {
	app.ensureFieldsSet()
	return &app.f.cfg.Custom
}

func (app *App[_]) SetTitle(title string) {
	app.ensureFieldsSet()
	app.f.title = title
}

func (app *App[_]) SetVersion(version string) {
	app.ensureFieldsSet()
	app.f.version = version
}

func (app *App[_]) SetDB(kind database.DBKind) {
	app.ensureFieldsSet()
	app.f.dbKind = kind
}

func (app *App[T]) Run(callback func(ctx *App[T]) error) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	app.ensureFieldsSet()
	err := app.runContext(ctx, callback)
	if err != nil {
		if app.f.hasLogger {
			app.f.logger.Err(err).Msg("shutting down")
		} else {
			fmt.Println(err.Error())
		}

		return err
	}

	app.f.logger.Info().Msg("shutting down")
	return nil
}

func (app *App[T]) runContext(ctx context.Context, callback func(ctx *App[T]) error) error {
	title := app.f.title
	if title == "" {
		title = "App"
	}

	version := app.f.version
	if version == "" {
		version = "0.0.1"
	}

	app.Context = ctx

	flag.BoolVar(&app.f.cfg.Common.Debug, "d", false, "")
	flag.BoolVar(&app.f.cfg.Common.Debug, "debug", false, "")
	flag.StringVar(&app.f.cfg.Common.ConfigFile, "c", "config.yml", "")
	flag.StringVar(&app.f.cfg.Common.ConfigFile, "config-file", "config.yml", "")
	flag.BoolVar(&app.f.cfg.Common.TracePerf, "t", false, "")
	flag.BoolVar(&app.f.cfg.Common.TracePerf, "trace-perf", false, "")

	flag.Usage = func() {
		fmt.Println(title + " v" + version + "\n" +
			"Usage:\n" +
			"\t-d, --debug: enable debug output\n" +
			"\t-c, --config-file: path to config file\n" +
			"\t-t, --trace-perf: enable perf tracing")
	}

	flag.Parse()

	err := app.loadConfig()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	app.makeLogger()

	err = app.initDB()
	if err != nil {
		return fmt.Errorf("initializing DB: %w", err)
	}
	defer app.shutdownDB()

	app.Log().Str("title", app.f.title).Str("version", app.f.version).Msg("app: running")
	return callback(app)
}

func (app *App[_]) loadConfig() error {
	f, err := os.Open(app.f.cfg.Common.ConfigFile)
	if err != nil {
		return fmt.Errorf("opening config file \"%v\": %w", app.f.cfg.Common.ConfigFile, err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&app.f.cfg)
	if err != nil {
		return fmt.Errorf("decoding YAML: %w", err)
	}

	return nil
}

func (app *App[T]) ensureFieldsSet() {
	if app.f == nil {
		app.f = &appFields[T]{}
	}
}
