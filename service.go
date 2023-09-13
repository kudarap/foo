package foo

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
)

// Service represents foo service.
type Service struct {
	repo   repository
	logger *slog.Logger
}

// NewService returns new foo service.
func NewService(r repository, l *slog.Logger) *Service {
	return &Service{repo: r, logger: l}
}

// FighterByID returns a fighter by id.
func (s *Service) FighterByID(ctx context.Context, sid string) (*Fighter, error) {
	// NOTE this is a just a demo logging and should use InfoContext enabling telemetry logs.
	s.logger.InfoContext(ctx, "getting foo fighter by id", "id", sid)

	id, err := uuid.Parse(sid)
	if err != nil {
		return nil, err
	}

	f, err := s.repo.Fighter(ctx, id)
	if err != nil {
		if errors.Is(err, ErrFighterNotFound) {
			return nil, ErrFighterNotFound.X(err)
		}
		return nil, fmt.Errorf("could not find fighter on repository: %s", err)
	}
	return f, nil
}

// repository manages storage operation for fighters.
type repository interface {
	Fighter(ctx context.Context, id uuid.UUID) (*Fighter, error)
}
