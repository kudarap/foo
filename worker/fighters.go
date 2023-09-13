package worker

import (
	"context"
	"log/slog"
	"math/rand"
	"time"
)

func FakeFighterConsumer(l *slog.Logger) JobHandler {
	return func(ctx context.Context, j Job) error {
		l.Info("fake-fighter consumer received", "job", string(j.Payload))
		// emulate process latency
		time.Sleep(time.Second*2 + time.Millisecond*time.Duration(rand.Int63n(3000)))
		return nil
	}
}
