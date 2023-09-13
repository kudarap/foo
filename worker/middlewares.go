package worker

import (
	"context"
	"log/slog"
	"time"
)

func LoggingMiddleware(logger *slog.Logger) func(next JobHandler) JobHandler {
	return func(next JobHandler) JobHandler {
		return func(ctx context.Context, job Job) error {
			start := time.Now()
			logger.InfoContext(ctx, "job received", "topic", job.Topic, "payload", string(job.Payload))

			if err := next(ctx, job); err != nil {
				logger.ErrorContext(ctx, "job handler", "err", err, "topic", job.Topic)
				return err
			}

			logger.InfoContext(ctx, "job success",
				"topic", job.Topic,
				"duration_ms", time.Now().Sub(start).Milliseconds())
			return nil
		}
	}
}
