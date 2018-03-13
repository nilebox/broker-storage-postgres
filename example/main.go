package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"os/signal"
	"syscall"

	"github.com/nilebox/broker-server/example/server"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	defaultAddr = ":8080"
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

	fs.Parse(os.Args[1:]) // nolint: gas

	app := server.ExampleServer{
		Addr: *addr,
	}
	return app.Run(ctx)
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
