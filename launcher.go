package launcher

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"
)

var _ Runnable = (*Launcher)(nil)

type Launcher struct {
	rs []Runnable
}

func NewLauncher(ls ...Runnable) Launcher {
	return Launcher{rs: ls}
}

func (l Launcher) Run(ctx context.Context) error {
	g, gCtx := errgroup.WithContext(ctx)

	for _, r := range l.rs {
		g.Go(func() error {
			return r.Run(gCtx)
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("an app in launcher has returned an error: %w", err)
	}

	return nil
}

func Launching(ls ...Runnable) Runnable {
	return NewLauncher(ls...)
}
