package main

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	transport "gitlab.crja72.ru/gospec/go5/rooms/internal/transport/grpc"
	"gitlab.crja72.ru/gospec/go5/rooms/internal/transport/grpc/proto"
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
)

func main() {
	ctx := context.Background()
	grpcServer, err := transport.NewServer(ctx, GRPCServerPort)

	if err != nil {
		panic(err)
	}

	conn, err := grpc.NewClient(
		"0.0.0.0:"+strconv.Itoa(GRPCServerPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		panic(err)
	}

	gwMux := runtime.NewServeMux()
	if err := proto.RegisterRoomsServiceHandler(ctx, gwMux, conn); err != nil {
		panic(err)
	}

	gwServer := &http.Server{
		Addr:    ":" + strconv.Itoa(RESTServerPort),
		Handler: gwMux,
	}

	graceCh := make(chan os.Signal, 1)
	signal.Notify(graceCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := grpcServer.Start(ctx); err != nil {
			panic(err)
		}
	}()

	go func() {
		if err := gwServer.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	<-graceCh

	if err := gwServer.Shutdown(ctx); err != nil {
		panic(err)
	}

	if err := grpcServer.Stop(ctx); err != nil {
		panic(err)
	}
}
