package main

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"
)

var _ Runnable = (*Launcher)(nil)

type Launcher struct {
	ls []Runnable
}

func NewLauncher(ls ...Runnable) Launcher {
	return Launcher{ls: ls}
}

func (l Launcher) Run(ctx context.Context) error {
	g, gCtx := errgroup.WithContext(ctx)

	for _, lnch := range l.ls {
		g.Go(func() error {
			return lnch.Run(gCtx)
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("an app in launcher has returned an error: %w", err)
	}

	return nil
}
