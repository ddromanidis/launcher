package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"
)

type RunnableChain struct {
	fs []func(Runnable) Runnable
}

func (rc RunnableChain) Retry(r uint, delay time.Duration) RunnableChain {
	f := func(next Runnable) Runnable {
		return F(func(ctx context.Context) error {
			if r == 0 {
				r = r - 1 // retry many many times - just a stupid hack for nights coding, need something better
			}

			var lastErr error

			for a := 0; a < int(r); a++ {
				err := next.Run(ctx)
				if err == nil {
					return nil // Success
				}

				if lastErr == nil {
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

func (rc RunnableChain) Replicas(count uint) RunnableChain {
	f := func(next Runnable) Runnable {
		return F(func(ctx context.Context) error {
			if count == 0 {
				count = 1 // Ensure at least one replica runs.
			}

			g, gCtx := errgroup.WithContext(ctx)

			for i := uint(0); i < count; i++ {
				g.Go(func() error {
					// return next.Launch(gCtx)
					return fmt.Errorf("replica %d returned error: %w", i+1, next.Run(gCtx))
				})
			}

			return g.Wait()
		})
	}
	rc.fs = append(rc.fs, f)
	return rc
}

func (rc RunnableChain) Recover() RunnableChain {
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

func (rc RunnableChain) OnCancel(cleanupFunc func()) RunnableChain {
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

func (rc RunnableChain) OnError(errFunc func(error)) RunnableChain {
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
func (rc RunnableChain) X(f func(Runnable) Runnable) RunnableChain {
	rc.fs = append(rc.fs, f)
	return rc
}

func (rc RunnableChain) Apply(l Runnable) Runnable {
	for i := len(rc.fs) - 1; i >= 0; i-- {
		l = rc.fs[i](l)
	}
	return l
}

func (rc RunnableChain) ApplyFrontToBack(l Runnable) Runnable {
	for _, f := range rc.fs {
		l = f(l)
	}
	return l
}
