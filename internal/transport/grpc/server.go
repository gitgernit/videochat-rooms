package grpc

import (
	"context"
	"fmt"
	"net"

	"gitlab.crja72.ru/gospec/go5/contracts/proto/rooms/go/proto"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type Server struct {
	grpcServer *grpc.Server
	listener   net.Listener
}

func NewServer(ctx context.Context, grpcHost string, grpcPort int) (*Server, error) {
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", grpcHost, grpcPort))
	if err != nil {
		return nil, err
	}

	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)
	proto.RegisterRoomsServiceServer(grpcServer, newRoomsService())

	return &Server{grpcServer, lis}, nil
}

func (s *Server) Start(ctx context.Context) error {
	eg := errgroup.Group{}

	eg.Go(func() error {
		return s.grpcServer.Serve(s.listener)
	})

	return eg.Wait()
}

func (s *Server) Stop(ctx context.Context) error {
	s.grpcServer.GracefulStop()

	return nil
}
