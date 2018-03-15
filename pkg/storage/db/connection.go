package db

import (
	"database/sql"
	"fmt"
	"github.com/pkg/errors"
)

const (
	psqlDriverName = "postgres"
)

func OpenConnection(config *PostgresConfig) (*sql.DB, error) {
	db, err := sql.Open(psqlDriverName, fmt.Sprintf("dbname=%s user=%s password=%s host=%s port=%d sslmode=disable",
		config.Database, config.Username, config.Password, config.Host, config.Port))
	if err != nil {
		return nil, errors.Wrap(err, "could not open database connection")
	}
	return db, nil
}
