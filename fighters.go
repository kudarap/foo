package foo

import (
	"errors"

	guuid "github.com/google/uuid"
)

var ErrFighterNotFound = errors.New("fighter not found")

type Fighter struct {
	ID        guuid.UUID
	FirstName string
	LastName  string
}
