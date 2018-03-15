package db

import (
	"fmt"
	"github.com/pkg/errors"

	// This loads the postgresql driver anonymously, and registers with database/sql under the hood
	"context"
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/mattes/migrate"
	"github.com/mattes/migrate/database"
	"github.com/mattes/migrate/database/postgres"
	_ "github.com/mattes/migrate/source/file"
	"github.com/nilebox/broker-storage-postgres/pkg/util"
	"time"
)

const (
	lockAttempts           = 10
	migrateLockedSleepTime = 1 * time.Second
	// TODO will it work when it will be used as a library in another project?
	defaultMigrationsDir = "sql"
)

func RunMigrations(ctx context.Context, conn *sql.DB) error {
	driver, err := postgres.WithInstance(conn, &postgres.Config{})
	if err != nil {
		return errors.Wrap(err, "could not create migrations database driver")
	}

	m, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s", defaultMigrationsDir), "postgres", driver)
	if err != nil {
		return errors.Wrap(err, "could not create migration")
	}

	for i := 1; i <= lockAttempts; i++ {
		err = m.Up()
		if err == nil || err == migrate.ErrNoChange {
			break
		}

		if err != database.ErrLocked {
			return errors.Wrap(err, "could not migrate")
		}

		if i == lockAttempts {
			return errors.New("could not perform database migration after attempting to acquire lock many times")
		}

		// TODO log warning
		//log.Warn("Could not grab database lock; unable to perform migration, sleeping...")
		err = util.Sleep(ctx, migrateLockedSleepTime)
		if err != nil {
			return errors.Wrap(err, "failure while sleeping for lock")
		}
	}

	return nil
}
