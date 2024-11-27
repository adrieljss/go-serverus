package db

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// Compares the error code to the pgerrcode library.
// If pgerr is nil returns false.
// Example usage:
//
//	pgErrorIs(err, pgerrcode.UniqueViolation)
func PgErrorIs(pgerr error, pgcode string) bool {
	if pgerr == nil {
		return false
	}
	return ToPgError(pgerr).Code == pgcode
}

// Transforms a normal error into a pgerror.
func ToPgError(pgerr error) *pgconn.PgError {
	if pgerr == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(pgerr, &pgErr) {
		return pgErr
	} else {
		return nil
	}
}
