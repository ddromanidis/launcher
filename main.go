package main

import (
	"context"
	"errors"
	"log"
	"time"
)

func main() {
	lb := RunnableChain{}.
		Replicas(3).
		Retry(1, 2*time.Second).
		Recover().
		X(func(next Runnable) Runnable {
			return F(func(ctx context.Context) error {
				err := next.Run(ctx)
				log.Println(err)
				return err
			})
		})

	l := lb.Apply(ll{})

	f := F(Kek)

	l = LaunchingWithAwesomeAnimation(l, f)

	if err := l.Run(context.Background()); err != nil {
		log.Println(err)
	}
}

var _ Runnable = (*ll)(nil)

type ll struct {
}

func (l ll) Run(ctx context.Context) error {
	return errors.New("test launch")
}

func Kek(context.Context) error {
	return nil
}
