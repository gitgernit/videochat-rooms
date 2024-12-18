package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"gitlab.crja72.ru/gospec/go5/rooms/internal/config"
	transport "gitlab.crja72.ru/gospec/go5/rooms/internal/transport/grpc"
	"gitlab.crja72.ru/gospec/go5/rooms/pkg/logger"
	"go.uber.org/zap"
)

var (
	serviceName = "rooms"
)

func main() {
	ctx := context.Background()
	mainLogger := logger.New(zap.DebugLevel, serviceName)
	ctx = context.WithValue(ctx, logger.LoggerKey, mainLogger)

	cfg, err := config.New()
	if err != nil {
		mainLogger.Fatal(ctx, err.Error())
		return
	}

	grpcServer, err := transport.NewServer(ctx, mainLogger, cfg.GRPCServerHost, cfg.RESTServerHost, cfg.GRPCServerPort, cfg.RESTServerPort)
	if err != nil {
		mainLogger.Fatal(ctx, err.Error())
		return
	}

	graceCh := make(chan os.Signal, 1)
	signal.Notify(graceCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := grpcServer.Start(ctx); err != nil {
			mainLogger.Error(ctx, err.Error())
		}
	}()

	<-graceCh

	if err := grpcServer.Stop(ctx); err != nil {
		mainLogger.Error(ctx, err.Error())
	}
}
