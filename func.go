package main

import "context"

// F is an adapter to allow the use of ordinary functions as Launchable applications.
// It implements the Launchable interface.
var _ Runnable = (*F)(nil)

type F func(ctx context.Context) error

func (f F) Run(ctx context.Context) error {
	return f(ctx)
}
