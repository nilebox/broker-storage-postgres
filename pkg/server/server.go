package server

import (
	"context"

	"github.com/nilebox/broker-server/pkg/api"
	"github.com/nilebox/broker-server/pkg/server"
	"github.com/nilebox/broker-server/pkg/stateful"
	"github.com/nilebox/broker-server/pkg/stateful/retry"
	"github.com/nilebox/broker-server/pkg/stateful/task"
	"github.com/nilebox/broker-storage-postgres/pkg/storage"
	"github.com/nilebox/broker-storage-postgres/pkg/storage/db"
	"go.uber.org/zap"
)

type PostgresBrokerServer struct {
	Addr string
}

func (b *PostgresBrokerServer) Run(ctx context.Context, log *zap.Logger, catalog *api.Catalog, broker task.Broker, config *db.PostgresConfig) (returnErr error) {
	// Create a storage for OSB
	strg, err := storage.NewPostgresStorage(ctx, config)
	if err != nil {
		return err
	}
	defer func() {
		err := strg.Close()
		if err != nil {
			returnErr = err
		}
	}()

	// Start a controller that runs actual broker implementation asynchronously
	retryController := retry.NewRetryController(strg, broker)
	retryController.Start(ctx)

	// Wrap a storage into a decorator that submits the broker task
	// after every instance update.
	// We'll use it for REST controller
	storageWithSubmitter := retryController.CreateStorageWithSubmitter()

	// Run a REST server
	controller := stateful.NewStatefulController(ctx, catalog, storageWithSubmitter)
	return server.Run(ctx, b.Addr, controller)
}
