package main

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/tmc/grpc-websocket-proxy/wsproxy"
	transport "gitlab.crja72.ru/gospec/go5/rooms/internal/transport/grpc"
	"gitlab.crja72.ru/gospec/go5/rooms/internal/transport/grpc/proto"
	"gitlab.crja72.ru/gospec/go5/rooms/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

var (
	GRPCServerPort = 50051
	RESTServerPort = 8081
	serviceName    = "rooms"
)

func main() {
	ctx := context.Background()
	mainLogger := logger.New(serviceName)
	ctx = context.WithValue(ctx, logger.LoggerKey, mainLogger)

	grpcServer, err := transport.NewServer(ctx, GRPCServerPort)

	if err != nil {
		mainLogger.Error(ctx, err.Error())
	}

	conn, err := grpc.NewClient(
		"0.0.0.0:"+strconv.Itoa(GRPCServerPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		mainLogger.Error(ctx, err.Error())
	}

	gwMux := runtime.NewServeMux()
	if err := proto.RegisterRoomsServiceHandler(ctx, gwMux, conn); err != nil {
		mainLogger.Error(ctx, err.Error())
	}

	gwServer := &http.Server{
		Addr:    ":" + strconv.Itoa(RESTServerPort),
		Handler: wsproxy.WebsocketProxy(gwMux),
	}

	graceCh := make(chan os.Signal, 1)
	signal.Notify(graceCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := grpcServer.Start(ctx); err != nil {
			mainLogger.Error(ctx, err.Error())
		}
	}()

	mainLogger.Info(ctx, "gRPC server started", zap.Int("port", GRPCServerPort))

	go func() {
		if err := gwServer.ListenAndServe(); err != nil {
			mainLogger.Error(ctx, err.Error())
		}
	}()

	mainLogger.Info(ctx, "Gateway server started", zap.Int("port", RESTServerPort))

	<-graceCh

	mainLogger.Info(ctx, "Shutting down")

	if err := gwServer.Shutdown(ctx); err != nil {
		mainLogger.Error(ctx, err.Error())
	}

	if err := grpcServer.Stop(ctx); err != nil {
		mainLogger.Error(ctx, err.Error())
	}

	mainLogger.Info(ctx, "Successfully shut down")
}
