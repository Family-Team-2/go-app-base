package appctx

import (
	"context"
	"time"

	"github.com/Family-Team-2/go-app-base/database"
	"github.com/rs/zerolog"
)

type AppCtx[T any] struct {
	context.Context
}

type AppCtxAny = AppCtx[any]

type appCtxKey struct{}

type appCtxContents[T any] struct {
	logger *zerolog.Logger
	db     *database.DB
	cfg    *T
}

func NewAppCtx[T any](ctx context.Context, cfg *T, logger *zerolog.Logger, db *database.DB) AppCtx[T] {
	ac := AppCtx[T]{
		Context: context.WithValue(ctx, appCtxKey{}, &appCtxContents[T]{
			logger: logger,
			db:     db,
			cfg:    cfg,
		}),
	}

	return ac
}

func FromContext[T any](ctx context.Context) *AppCtx[T] {
	ac, ok := ctx.(*AppCtx[T])
	if !ok || ac == nil {
		return &AppCtx[T]{
			ctx,
		}
	} else {
		return ac
	}
}

func (ac *AppCtx[T]) Logger() *zerolog.Logger {
	return contents[T](ac).logger
}

func (ac *AppCtx[T]) Log() *zerolog.Event {
	return contents[T](ac).logger.Info()
}

func (ac *AppCtx[T]) Debug() *zerolog.Event {
	return contents[T](ac).logger.Debug()
}

func (ac *AppCtx[T]) Error() *zerolog.Event {
	return contents[T](ac).logger.Error()
}

func (ac *AppCtx[T]) Cfg() *T {
	return contents[T](ac).cfg
}

func (ac *AppCtx[T]) DBC(ctx context.Context) *database.DB {
	return &database.DB{
		DB: contents[T](ac).db.WithContext(ctx),
	}
}

func (ac *AppCtx[T]) DB() *database.DB {
	return ac.DBC(ac)
}

func (ac *AppCtx[T]) WithValue(key, val any) *AppCtx[T] {
	return &AppCtx[T]{
		Context: context.WithValue(ac, key, val),
	}
}

func (ac *AppCtx[T]) WithTimeout(d time.Duration) (*AppCtx[T], func()) {
	ctx, cancel := context.WithTimeout(ac, d)
	return &AppCtx[T]{
		Context: ctx,
	}, cancel
}

func contents[T any](ctx context.Context) *appCtxContents[T] {
	c, ok := ctx.Value(appCtxKey{}).(*appCtxContents[T])
	if !ok || c == nil {
		panic("missing contents from app context")
	}

	return c
}
