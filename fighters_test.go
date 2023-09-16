package foo_test

import (
	"context"
	"log/slog"
	"os"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/kudarap/foo"
)

func TestService_FighterByID(t *testing.T) {
	tests := []struct {
		name string
		// dependencies
		repo *mockRepo
		// params
		fighterUUID string
		// returns
		want    *foo.Fighter
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			"place-holder-test",
			&mockRepo{
				FighterFn: func(ctx context.Context, id uuid.UUID) (*foo.Fighter, error) {
					return &foo.Fighter{
						ID:        uuid.MustParse("b41c7709-04e3-4c48-b233-34e6838d9140"),
						FirstName: "justine",
						LastName:  "jimenez",
					}, nil
				}},
			"b41c7709-04e3-4c48-b233-34e6838d9140",
			&foo.Fighter{
				ID:        uuid.MustParse("b41c7709-04e3-4c48-b233-34e6838d9140"),
				FirstName: "justine",
				LastName:  "jimenez",
			},
			false,
		},
		// TODO: Add test cases.
		// TODO: Add negative test cases.
		// TODO: Add fatal repo test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := slog.New(slog.NewTextHandler(os.Stdout, nil))
			svc := foo.NewService(tt.repo, l)
			ctx := context.Background()
			got, err := svc.FighterByID(ctx, tt.fighterUUID)
			if (err != nil) != tt.wantErr {
				t.Errorf("FighterByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FighterByID() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type mockRepo struct {
	FighterFn func(ctx context.Context, id uuid.UUID) (*foo.Fighter, error)
	ArcherFn  func(ctx context.Context, id uuid.UUID) (*foo.Archer, error)
}

type mockArcherRepo struct {
	ArcherFn func(ctx context.Context, id uuid.UUID) (*foo.Archer, error)
}

// Archer implements foo.repository.
func (m *mockRepo) Archer(ctx context.Context, id uuid.UUID) (*foo.Archer, error) {
	return m.ArcherFn(ctx, id)
}

func (m *mockRepo) Fighter(ctx context.Context, id uuid.UUID) (*foo.Fighter, error) {
	return m.FighterFn(ctx, id)
}
