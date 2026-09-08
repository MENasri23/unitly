package db

import (
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// IsUniqueViolation reports whether err is a unique-constraint/index
// violation on the given constraint or index name (pass "" to match any
// unique violation). Repositories use this to translate a raw Postgres
// error into the right domain sentinel, e.g.:
//
//	if db.IsUniqueViolation(err, "uq_key") {
//	    return ErrEntityTaken
//	}
func IsUniqueViolation(err error, constraintName string) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != pgerrcode.UniqueViolation {
		return false
	}
	return constraintName == "" || pgErr.ConstraintName == constraintName
}

func IsNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
