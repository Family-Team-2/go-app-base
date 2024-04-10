package appctx

import (
	"context"
	"time"

	"github.com/Family-Team-2/go-app-base/database"
	"github.com/rs/zerolog"
)

type AppCtx struct {
	context.Context
}

type appCtxKey struct{}

type appCtxContents struct {
	logger *zerolog.Logger
	db     *database.DB
}

func NewAppCtx(ctx context.Context, logger *zerolog.Logger, db *database.DB) AppCtx {
	ac := AppCtx{
		context.WithValue(ctx, appCtxKey{}, &appCtxContents{
			logger: logger,
			db:     db,
		}),
	}

	return ac
}

func FromContext(ctx context.Context) *AppCtx {
	ac, ok := ctx.(*AppCtx)
	if !ok || ac == nil {
		return &AppCtx{
			ctx,
		}
	} else {
		return ac
	}
}

func (ac *AppCtx) Logger() *zerolog.Logger {
	return contents(ac).logger
}

func (ac *AppCtx) Log() *zerolog.Event {
	return contents(ac).logger.Info()
}

func (ac *AppCtx) Debug() *zerolog.Event {
	return contents(ac).logger.Debug()
}

func (ac *AppCtx) Error() *zerolog.Event {
	return contents(ac).logger.Error()
}

func (ac *AppCtx) DBC(ctx context.Context) *database.DB {
	return &database.DB{
		DB: contents(ac).db.WithContext(ctx),
	}
}

func (ac *AppCtx) DB() *database.DB {
	return ac.DBC(ac)
}

func (ac *AppCtx) WithValue(key, val any) *AppCtx {
	return &AppCtx{
		Context: context.WithValue(ac, key, val),
	}
}

func (ac *AppCtx) WithTimeout(d time.Duration) (*AppCtx, func()) {
	ctx, cancel := context.WithTimeout(ac, d)
	return &AppCtx{
		Context: ctx,
	}, cancel
}

func contents(ctx context.Context) *appCtxContents {
	c, ok := ctx.Value(appCtxKey{}).(*appCtxContents)
	if !ok || c == nil {
		panic("missing contents from app context")
	}

	return c
}
