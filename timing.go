package goapp

import (
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type timingContext struct {
	name     string
	children []*timingContext

	isRoot   bool
	start    time.Time
	duration time.Duration
}

type timingContextKey struct{}

func (app *App[T]) Perf(names ...string) (newApp *App[T], done func()) {
	if !app.f.cfg.Common.TracePerf {
		return app, func() {}
	}

	tc := &timingContext{
		name:  strings.Join(names, "/"),
		start: time.Now(),
	}

	app.f.timingMu.Lock()
	defer app.f.timingMu.Unlock()

	parent, ok := app.Value(timingContextKey{}).(*timingContext)
	if !ok {
		tc.isRoot = true
	} else {
		parent.children = append(parent.children, tc)
	}

	done = func() {
		app.f.timingMu.Lock()
		defer app.f.timingMu.Unlock()

		tc.duration = time.Since(tc.start)
		if tc.isRoot && app.Logger() != nil {
			tc.Print(app.Logger())
		}
	}

	return app.WithValue(timingContextKey{}, tc), done
}

func (tc *timingContext) Print(logger *zerolog.Logger) {
	s := "timing: " + tc.name

	s2 := tc.internalPrint(0)
	if s2 != "" {
		s += "\n" + s2 + "total " + tc.duration.String()
	} else {
		s += ": total " + tc.duration.String()
	}

	logger.Debug().Msg(s)
}

func (tc *timingContext) internalPrint(padding int) string {
	if len(tc.children) == 0 {
		return ""
	}

	s := strings.Repeat(" ", padding)

	for _, child := range tc.children {
		s += "- " + child.name + ": " + child.duration.String() + "\n"
		if len(child.children) > 0 {
			s += child.internalPrint(padding + 2)
		}
	}

	return s
}
