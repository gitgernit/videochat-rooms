package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/tmc/grpc-websocket-proxy/wsproxy"
	"gitlab.crja72.ru/gospec/go5/contracts/proto/rooms/go/proto"
	"gitlab.crja72.ru/gospec/go5/rooms/internal/config"
	transport "gitlab.crja72.ru/gospec/go5/rooms/internal/transport/grpc"
	"gitlab.crja72.ru/gospec/go5/rooms/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	serviceName = "rooms"
)

func main() {
	ctx := context.Background()
	mainLogger := logger.New(zap.InfoLevel, serviceName)
	ctx = context.WithValue(ctx, logger.LoggerKey, mainLogger)

	cfg, err := config.New()
	if err != nil {
		mainLogger.Fatal(ctx, err.Error())
		return
	}
	grpcServer, err := transport.NewServer(ctx, cfg.GRPCServerHost, cfg.GRPCServerPort)

	if err != nil {
		mainLogger.Fatal(ctx, err.Error())
		return
	}

	conn, err := grpc.NewClient(
		fmt.Sprintf("%s:%d", cfg.GRPCServerHost, cfg.GRPCServerPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		mainLogger.Fatal(ctx, err.Error())
		return
	}

	gwMux := runtime.NewServeMux()
	if err := proto.RegisterRoomsServiceHandler(ctx, gwMux, conn); err != nil {
		mainLogger.Fatal(ctx, err.Error())
		return
	}

	gwServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.RESTServerHost, cfg.RESTServerPort),
		Handler: wsproxy.WebsocketProxy(gwMux),
	}

	graceCh := make(chan os.Signal, 1)
	signal.Notify(graceCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := grpcServer.Start(ctx); err != nil {
			mainLogger.Error(ctx, err.Error())
		}
	}()

	mainLogger.Info(ctx, "gRPC server started", zap.Int("port", cfg.GRPCServerPort))

	go func() {
		if err := gwServer.ListenAndServe(); err != nil {
			mainLogger.Error(ctx, err.Error())
			return
		}
	}()

	mainLogger.Info(ctx, "Gateway server started", zap.Int("port", cfg.RESTServerPort))

	<-graceCh

	mainLogger.Info(ctx, "Shutting down")

	if err := gwServer.Shutdown(ctx); err != nil {
		mainLogger.Error(ctx, err.Error())
		return
	}

	if err := grpcServer.Stop(ctx); err != nil {
		mainLogger.Error(ctx, err.Error())
		return
	}

	mainLogger.Info(ctx, "Successfully shut down")
}
