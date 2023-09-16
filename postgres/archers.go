package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kudarap/foo"
)

func (c *Client) Archer(ctx context.Context, id uuid.UUID) (*foo.Archer, error) {
	var archer foo.Archer
	archer.ID = id
	err := c.db.
		QueryRow(ctx, `SELECT id, first_name, last_name FROM archers WHERE id=$1`, id.String()).
		Scan(&archer.ID, &archer.FirstName, &archer.LastName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, foo.ErrArcherNotFound
		}
		return nil, err
	}

	return &archer, nil
}
