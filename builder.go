package launcher

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/sync/errgroup"
)

type LauncherChain struct {
	fs []func(Runnable) Runnable
	l  *slog.Logger
}

func NewLauncherChain() *LauncherChain {
	return &LauncherChain{}
}

func (rc *LauncherChain) Retry(r uint, delay time.Duration) *LauncherChain {
	f := func(next Runnable) Runnable {
		return F(func(ctx context.Context) error {
			var lastErr error

			// r == 0 means loop will be infinite
			for a := 0; r == 0 || a < int(r); a++ {
				err := next.Run(ctx)
				if err == nil {
					return nil // Success
				}

				if rc.l != nil {
					rc.l.Error("retrying with error", "attempt", a, "error", err)
				}

				if lastErr == nil || r == 0 {
					lastErr = err
				} else {
					lastErr = errors.Join(lastErr, err)
				}

				select {
				case <-time.After(delay):
				case <-ctx.Done():
					if a > 0 { // If attempts were made to restart an app
						return errors.Join(
							ctx.Err(),
							fmt.Errorf("context cancelled during retry delay: %w", lastErr),
						)
					}
					return ctx.Err()
				}

			}
			return fmt.Errorf("failed after %d retries: %w", r, lastErr)
		})
	}
	rc.fs = append(rc.fs, f)
	return rc
}

func (rc *LauncherChain) Replicas(count uint) *LauncherChain {
	f := func(next Runnable) Runnable {
		return F(func(ctx context.Context) error {
			if count == 0 {
				count = 1 // Ensure at least one replica runs.
			}

			g, gCtx := errgroup.WithContext(ctx)

			for i := uint(0); i < count; i++ {
				i := i
				g.Go(func() error {
					if err := next.Run(ctx); err != nil {
						return fmt.Errorf("replica %d returned error: %w", i+1, next.Run(gCtx))
					}

					if rc.l != nil {
						rc.l.Info("replica finished without error", "replica_n", i+1)
					}
					return nil
				})
			}

			return g.Wait()
		})
	}
	rc.fs = append(rc.fs, f)
	return rc
}

func (rc *LauncherChain) Recover() *LauncherChain {
	f := func(next Runnable) Runnable {
		return F(func(ctx context.Context) (err error) {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("panic recovered: %v", r)
				}
			}()
			return next.Run(ctx)
		})
	}

	rc.fs = append(rc.fs, f)
	return rc
}

func (rc *LauncherChain) OnCancel(cleanupFunc func()) *LauncherChain {
	f := func(next Runnable) Runnable {
		return F(func(ctx context.Context) error {
			err := next.Run(ctx)
			if err != nil && errors.Is(err, context.Canceled) {
				// appName := GetAppName(next)
				// fmt.Printf("Running OnCancel for %s due to context cancellation\n", appName)
				cleanupFunc()
			}
			return err
		})
	}
	rc.fs = append(rc.fs, f)
	return rc
}

func (rc *LauncherChain) OnError(errFunc func(error)) *LauncherChain {
	f := func(next Runnable) Runnable {
		return F(func(ctx context.Context) error {
			err := next.Run(ctx)
			if err != nil {
				errFunc(err)
			}
			return err
		})
	}
	rc.fs = append(rc.fs, f)
	return rc
}

// X method allows you to extend LaunchableBuilder with methods of your own
func (rc *LauncherChain) X(f func(Runnable) Runnable) *LauncherChain {
	rc.fs = append(rc.fs, f)
	return rc
}

func (rc *LauncherChain) WithLogger(l *slog.Logger) *LauncherChain {
	rc.l = l
	return rc
}

func (rc *LauncherChain) Apply(l Runnable) Runnable {
	for i := len(rc.fs) - 1; i >= 0; i-- {
		l = rc.fs[i](l)
	}
	return l
}
