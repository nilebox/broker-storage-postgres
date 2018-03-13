package server

import (
	"context"

	"github.com/nilebox/broker-server/pkg/server"
	"github.com/nilebox/broker-server/pkg/stateful"
	"github.com/nilebox/broker-server/pkg/stateful/retry"
	examplebroker "github.com/nilebox/broker-storage-postgres/example/broker"
	"github.com/nilebox/broker-storage-postgres/pkg/storage"
	"go.uber.org/zap"
)

type ExampleServer struct {
	Addr string
}

func (b *ExampleServer) Run(ctx context.Context) (returnErr error) {
	log := ctx.Value("log").(*zap.Logger)
	_ = log

	// Create a storage for OSB
	storage := storage.NewPostgresStorage()

	// Start a controller that runs actual broker implementation asynchronously
	broker, err := examplebroker.NewExampleBroker(log)
	if err != nil {
		panic("Broker creation failed: " + err.Error())
	}
	retryController := retry.NewRetryController(storage, broker)
	retryController.Start(ctx)

	// Run a REST server
	controller := stateful.NewStatefulController(ctx, examplebroker.Catalog(), storage)
	return server.Run(ctx, b.Addr, controller)
}
