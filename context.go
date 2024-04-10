package main

import (
	"context"
	"time"
)

func (c *App[T]) WithValue(key, val any) *App[T] {
	newApp := c.clone()
	newApp.Context = context.WithValue(newApp.Context, key, val)
	return newApp
}

func (c *App[T]) WithTimeout(d time.Duration) (*App[T], func()) {
	var cancel func()
	newApp := c.clone()
	newApp.Context, cancel = context.WithTimeout(newApp.Context, d)
	return newApp, cancel
}

func (c *App[T]) WithCancel() (*App[T], func()) {
	var cancel func()
	newApp := c.clone()
	newApp.Context, cancel = context.WithCancel(newApp.Context)
	return newApp, cancel
}
