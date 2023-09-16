package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kudarap/foo"
)

func (c *Client) Fighter(ctx context.Context, id uuid.UUID) (*foo.Fighter, error) {
	var fighter foo.Fighter
	fighter.ID = id
	err := c.db.
		QueryRow(ctx, `SELECT id, first_name, last_name FROM fighters WHERE id=$1`, id.String()).
		Scan(&fighter.ID, &fighter.FirstName, &fighter.LastName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, foo.ErrFighterNotFound
		}
		return nil, err
	}

	return &fighter, nil
}
