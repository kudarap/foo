package telemetry

import (
	"context"

	"github.com/kudarap/foo/worker"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// TraceWorker traces worker job handler results.
func TraceWorker(next worker.JobHandler) worker.JobHandler {
	return func(ctx context.Context, job worker.Job) error {
		ctx, span := otel.Tracer("worker").Start(ctx, job.Topic)
		defer span.End()
		span.SetAttributes(
			attribute.String("topic", job.Topic),
			attribute.String("payload", string(job.Payload)),
		)

		if err := next(ctx, job); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}
		return nil
	}
}
