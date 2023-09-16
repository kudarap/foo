package foo

import (
	guuid "github.com/google/uuid"
	"github.com/kudarap/foo/xerror"
)

var ErrArcherNotFound = xerror.Error("not_found")

type Archer struct {
	ID        guuid.UUID
	FirstName string
	LastName  string
}
