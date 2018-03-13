package server

import (
	"context"

	"github.com/nilebox/broker-server/pkg/server"
	"github.com/nilebox/broker-server/pkg/stateful"
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

	s := storage.NewPostgresStorage()
	c := stateful.NewStatefulController(ctx, examplebroker.Catalog(), s)

	// TODO start a scheduler and watch dog

	return server.Run(ctx, b.Addr, c)
}
