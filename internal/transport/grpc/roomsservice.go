package grpc

import (
	"gitlab.crja72.ru/gospec/go5/rooms/internal/transport/grpc/proto"
	"io"
)

type roomsService struct {
	proto.UnimplementedRoomsServiceServer
}

func newRoomsService() *roomsService {
	return &roomsService{}
}

func (s *roomsService) PingPong(stream proto.RoomsService_PingPongServer) error {
	for {
		msg, err := stream.Recv()

		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		counter := msg.GetCounter()
		counter += 1

		response := proto.Pong{Counter: counter}

		if err := stream.Send(&response); err != nil {
			return err
		}
	}
}
