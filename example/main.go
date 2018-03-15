package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"os/signal"
	"syscall"

	examplebroker "github.com/nilebox/broker-storage-postgres/example/broker"
	"github.com/nilebox/broker-storage-postgres/pkg/server"
	"github.com/nilebox/broker-storage-postgres/pkg/storage/db"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	defaultAddr = ":8080"

	defaultPgDatabase = "osb"
	defaultPgHost     = "localhost"
	defaultPgPort     = 5432
	defaultPgUsername = "test"
	defaultPgPassword = "test"
)

func main() {
	if err := run(); err != nil && err != context.Canceled && err != context.DeadlineExceeded {
		fmt.Fprintf(os.Stderr, "%+v", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	cancelOnInterrupt(ctx, cancelFunc)

	log := initializeLogger()
	defer log.Sync()
	ctx = context.WithValue(ctx, "log", log)

	return runWithContext(ctx)
}

func runWithContext(ctx context.Context) error {
	log := ctx.Value("log").(*zap.Logger)
	_ = log

	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	addr := fs.String("addr", defaultAddr, "Address to listen on")
	pgDatabase := fs.String("pg_database", defaultPgDatabase, "Database to use for PostgreSQL")
	pgHost := fs.String("pg_host", defaultPgHost, "Hostname to use for PostgreSQL")
	pgPort := fs.Int("pg_port", defaultPgPort, "Port to use for PostgreSQL")
	pgUsername := fs.String("pg_username", defaultPgUsername, "Username to use for PostgreSQL")
	pgPassword := fs.String("pg_password", defaultPgPassword, "Password to use for PostgreSQL")

	fs.Parse(os.Args[1:]) // nolint: gas

	broker, err := examplebroker.NewExampleBroker(log)
	if err != nil {
		// TODO log error and exit
		panic(err.Error())
	}

	app := server.PostgresBrokerServer{
		Addr: *addr,
	}
	pgConfig := db.PostgresConfig{
		Database: *pgDatabase,
		Host:     *pgHost,
		Port:     *pgPort,
		Username: *pgUsername,
		Password: *pgPassword,
	}
	return app.Run(ctx, examplebroker.Catalog(), broker, &pgConfig)
}

func initializeLogger() *zap.Logger {
	return zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
			zapcore.Lock(zapcore.AddSync(os.Stdout)),
			zap.InfoLevel,
		),
		zap.AddCaller(),
		zap.Fields(),
	)
}

// cancelOnInterrupt calls f when os.Interrupt or SIGTERM is received.
// It ignores subsequent interrupts on purpose - program should exit correctly after the first signal.
func cancelOnInterrupt(ctx context.Context, f context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case <-ctx.Done():
		case <-c:
			f()
		}
	}()
}
