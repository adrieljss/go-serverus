package db

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// use the pgerrcode library, if pgerr is nil, returns false also
//
// example usage
//
//	pgErrorIs(err, pgerrcode.UniqueViolation)
func PgErrorIs(pgerr error, pgcode string) bool {
	if pgerr == nil {
		return false
	}
	return ToPgError(pgerr).Code == pgcode
}

// transform a normal error into pgError
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
