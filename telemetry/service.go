package telemetry

import (
	"context"

	"github.com/kudarap/foo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type FooService struct {
	*foo.Service
	tracerName string
}

func (s *FooService) FighterByID(ctx context.Context, id string) (*foo.Fighter, error) {
	ctx, span := otel.Tracer(s.tracerName).Start(ctx, "fooservice.FighterByID")
	defer span.End()
	span.SetAttributes(attribute.String("id", id))

	f, err := s.Service.FighterByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return f, nil
}

func TraceFooService(s *foo.Service) *FooService {
	return &FooService{s, "foo-service"}
}
