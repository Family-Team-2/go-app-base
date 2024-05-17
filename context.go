package goapp

import (
	"context"
	"time"
)

func (app *App[T]) WithValue(key, val any) *App[T] {
	return app.cloneWithContext(context.WithValue(app.Context, key, val))
}

func (app *App[T]) WithTimeout(d time.Duration) (newApp *App[T], done func()) {
	ctx, cancel := context.WithTimeout(app.Context, d)
	return app.cloneWithContext(ctx), cancel
}

func (app *App[T]) WithCancel() (newApp *App[T], done func()) {
	ctx, cancel := context.WithCancel(app.Context)
	return app.cloneWithContext(ctx), cancel
}

func (app *App[T]) WithContext(ctx context.Context) *App[T] {
	return app.cloneWithContext(ctx)
}

func (app *App[T]) cloneWithContext(ctx context.Context) *App[T] {
	return &App[T]{
		Context: ctx,
		f:       app.f,
	}
}
