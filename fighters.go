package foo

import (
	guuid "github.com/google/uuid"
	"github.com/kudarap/foo/xerror"
)

var ErrFighterNotFound = xerror.Error("not_found")

type Fighter struct {
	ID        guuid.UUID
	FirstName string
	LastName  string
}
