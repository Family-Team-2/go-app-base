package timing

import (
	"strings"
	"time"

	"github.com/Family-Team-2/go-app-base/appctx"
	"github.com/rs/zerolog"
)

type TimingContext struct {
	name     string
	children []*TimingContext

	isRoot   bool
	start    time.Time
	duration time.Duration
}

type timingContextKey struct{}

func Start(ctx *appctx.AppCtxAny, names ...string) (*appctx.AppCtxAny, func()) {
	tc := &TimingContext{
		name:  strings.Join(names, "/"),
		start: time.Now(),
	}

	parent, ok := ctx.Value(timingContextKey{}).(*TimingContext)
	if !ok {
		tc.isRoot = true
	} else {
		parent.children = append(parent.children, tc)
	}

	done := func() {
		tc.duration = time.Since(tc.start)
		if tc.isRoot {
			tc.Print(ctx.Logger())
		}
	}

	return ctx.WithValue(timingContextKey{}, tc), done
}

func (tc *TimingContext) Print(logger *zerolog.Logger) {
	s := "timing: " + tc.name

	s2 := tc.internalPrint(0)
	if s2 != "" {
		s += "\n" + s2 + "total " + tc.duration.String()
	} else {
		s += ": total " + tc.duration.String()
	}

	logger.Debug().Msg(s)
}

func (tc *TimingContext) internalPrint(padding int) string {
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
