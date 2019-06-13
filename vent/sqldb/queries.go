package sqldb

import (
	"github.com/jmoiron/sqlx"
)

type Queries struct {
	LastBlockHeight *sqlx.NamedStmt
	SetBlockHeight  string
}
