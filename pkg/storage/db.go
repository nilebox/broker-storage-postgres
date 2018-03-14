package storage

import (
	"database/sql"
	"io"

	"github.com/lib/pq"
	"github.com/pkg/errors"
)

const (
	PgUniqueViolation = "23505"
)

func IsUniqueViolation(err error) bool {
	cause := errors.Cause(err)
	if pgerr, ok := cause.(*pq.Error); ok {
		if pgerr.Code == PgUniqueViolation {
			return true
		}
	}
	return false
}

func InTransaction(db *sql.DB, q func(*sql.Tx) (interface{}, error)) (result interface{}, returnErr error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "could not begin transaction")
	}

	defer func() {
		if returnErr != nil {
			tx.Rollback()
			return
		}
		returnErr = errors.Wrap(tx.Commit(), "could not commit transaction")
	}()

	return q(tx)
}

func Close(closable io.Closer, err *error) {
	closeErr := closable.Close()
	if *err == nil {
		*err = errors.Wrap(closeErr, "error while closing closer")
	}
}
