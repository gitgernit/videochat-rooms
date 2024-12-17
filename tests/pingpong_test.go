package tests

import (
	"context"
	"gitlab.crja72.ru/gospec/go5/contracts/proto/rooms/go/proto"
	"gitlab.crja72.ru/gospec/go5/rooms/internal/infrastructure/rooms/repositories/memory"
	transport "gitlab.crja72.ru/gospec/go5/rooms/internal/transport/grpc"
	"gitlab.crja72.ru/gospec/go5/rooms/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"log"
	"net"
	"testing"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func init() {
	mainLogger := logger.New(zap.DebugLevel, "test")

	lis = bufconn.Listen(bufSize)
	var opts []grpc.ServerOption

	repository := memory.NewRepository()

	grpcServer := grpc.NewServer(opts...)
	proto.RegisterRoomsServiceServer(grpcServer, transport.NewRoomsService(mainLogger, repository))
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestPingPongGrpc(t *testing.T) {
	ctx := context.Background()
	//conn, err := grpc.NewClient("bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := proto.NewRoomsServiceClient(conn)
	stream, err := client.PingPong(ctx)
	if err != nil {
		t.Fatal(err)
	}

	var counter uint32 = 0

	for counter < 100 {
		ping := proto.Ping{Counter: counter}
		req := &proto.Ping{Counter: ping.Counter}
		if err := stream.Send(req); err != nil {
			t.Fatalf("failed to send a ping request: %v", err)
		}

		pong, err := stream.Recv()
		if err != nil {
			t.Fatal(err)
		}

		counter = pong.Counter
	}
}
