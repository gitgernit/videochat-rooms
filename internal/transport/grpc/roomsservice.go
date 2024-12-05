package grpc

import (
	"gitlab.crja72.ru/gospec/go5/rooms/internal/domain/pingpong"
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
	interactor := pingpong.PingInteractor{}

	for {
		msg, err := stream.Recv()

		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		counter := msg.GetCounter()
		ping := pingpong.Ping{Counter: counter}

		pong, err := interactor.Ping(ping)
		if err != nil {
			return err
		}

		response := proto.Pong{Counter: pong.Counter}

		if err := stream.Send(&response); err != nil {
			return err
		}
	}
}
