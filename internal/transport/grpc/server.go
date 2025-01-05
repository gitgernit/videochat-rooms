package grpc

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/gitgernit/videochat-rooms/internal/infrastructure/rooms/repositories/memory"
	"github.com/gitgernit/videochat-rooms/pkg/logger"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"github.com/tmc/grpc-websocket-proxy/wsproxy"

	"github.com/gitgernit/videochat-contracts/proto/rooms/go/proto"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Server struct {
	grpcServer   *grpc.Server
	grpcListener net.Listener
	gwServer     *http.Server
}

func NewServer(
	ctx context.Context,
	logger logger.Logger,
	incomingRoomsChannel chan string,
	grpcHost,
	gwHost string,
	grpcPort,
	gwPort int,
) (*Server, error) {
	grpcLis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", grpcHost, grpcPort))
	if err != nil {
		return nil, err
	}

	var opts []grpc.ServerOption

	repository := memory.NewRepository()

	grpcServer := grpc.NewServer(opts...)
	proto.RegisterRoomsServiceServer(grpcServer, NewRoomsService(logger, repository, incomingRoomsChannel))

	gwMux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(RoomsHeaderMatcher),
	)
	wsMux := wsproxy.WebsocketProxy(gwMux,
		wsproxy.WithRequestMutator(WebsocketParamMutator),
		wsproxy.WithForwardedHeaders(
			func(header string) bool {
				_, ok := RoomsHeaderMatcher(header)
				return ok
			},
		),
	)

	conn, err := grpc.NewClient(
		fmt.Sprintf("%s:%d", grpcHost, grpcPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	if err := proto.RegisterRoomsServiceHandler(ctx, gwMux, conn); err != nil {
		return nil, err
	}

	corsMux := cors.New(cors.Options{
		AllowOriginFunc:  func(origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"ACCEPT", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}).Handler(wsMux)

	gwServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", gwHost, gwPort),
		Handler: corsMux,
	}

	return &Server{grpcServer, grpcLis, gwServer}, nil
}

func (s *Server) Start(ctx context.Context) error {
	l := logger.GetLoggerFromCtx(ctx)
	eg := errgroup.Group{}

	eg.Go(func() error {
		l.Info(ctx, "grpc: server start")
		return s.grpcServer.Serve(s.grpcListener)
	})

	eg.Go(func() error {
		l.Info(ctx, "gateway: server start")
		return s.gwServer.ListenAndServe()
	})

	return eg.Wait()
}

func (s *Server) Stop(ctx context.Context) error {
	l := logger.GetLoggerFromCtx(ctx)
	var err error
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		s.grpcServer.GracefulStop()
		l.Info(ctx, "grpc: server stopped")
	}()

	go func() {
		defer wg.Done()
		err = s.gwServer.Shutdown(ctx)
		l.Info(ctx, "gateway: server stopped")
	}()

	wg.Wait()
	return err
}
