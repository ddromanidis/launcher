package main

import (
	"context"
	"log"
	"log/slog"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"github.com/ddromanidis/launcher"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	lb := launcher.NewLauncherChain().
		Replicas(3).
		Retry(0, 1*time.Second).
		Recover().
		WithLogger(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	l := lb.Apply(ll{})

	f := launcher.F(Kek)

	l = launcher.Launching(l, f)

	if err := l.Run(ctx); err != nil {
		log.Println(err)
	}
}

var _ launcher.Runnable = (*ll)(nil)

type ll struct {
}

func (l ll) Run(ctx context.Context) error {
	if rand.Int31n(10) < 8 {
		panic("test panic")
	}
	return nil
}

func Kek(context.Context) error {
	return nil
}
