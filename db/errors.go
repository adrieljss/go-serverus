package db

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// returns constraint name if true, else if false retun ""
//
// examples
//
//	`users_user_id_key`
//	`users_email_key`
func IsDuplicateKeyError(pgerr error) string {
	if pgerr == nil {
		return ""
	}
	var pgErr *pgconn.PgError
	errors.As(pgerr, &pgErr)
	if pgErr.Code == "23505" {
		return pgErr.ConstraintName
	}
	return ""
}
