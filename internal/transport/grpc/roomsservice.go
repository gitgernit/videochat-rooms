package grpc

import (
	"context"
	"io"
	"net/http"
	"slices"

	"gitlab.crja72.ru/gospec/go5/contracts/proto/rooms/go/proto"
	"gitlab.crja72.ru/gospec/go5/rooms/internal/domain/pingpong"
	"gitlab.crja72.ru/gospec/go5/rooms/internal/domain/rooms"
	"gitlab.crja72.ru/gospec/go5/rooms/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	usernameMetadata = "username"
	roomIDMetadata   = "room_id"
)

func RoomsHeaderMatcher(key string) (string, bool) {
	switch key {
	case "Username":
		return usernameMetadata, true
	case "Room-Id":
		return roomIDMetadata, true
	default:
		return key, false
	}
}

func WebsocketParamMutator(incoming *http.Request, outgoing *http.Request) *http.Request {
	params := incoming.URL.Query()

	for key, values := range params {
		for _, value := range values {
			outgoing.Header.Add(key, value)
		}
	}

	return outgoing
}

type roomsService struct {
	proto.UnimplementedRoomsServiceServer
	logger     logger.Logger
	repository rooms.Repository
	Users      map[grpc.ServerStream]rooms.User
}

func newRoomsService(logger logger.Logger, repository rooms.Repository) *roomsService {
	return &roomsService{
		logger:     logger,
		repository: repository,
		Users:      make(map[grpc.ServerStream]rooms.User),
	}
}

func (s *roomsService) PingPong(stream proto.RoomsService_PingPongServer) error {
	interactor := pingpong.Interactor{}

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

func (s *roomsService) CreateRoom(ctx context.Context, req *proto.CreateRoomRequest) (*proto.CreateRoomResponse, error) {
	interactor := rooms.NewInteractor(s.logger, s.repository)

	id, err := interactor.CreateRoom()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &proto.CreateRoomResponse{Id: id}, nil
}

func (s *roomsService) JoinRoom(stream proto.RoomsService_JoinRoomServer) error {
	ctx := stream.Context()
	interactor := rooms.NewInteractor(s.logger, s.repository)

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.InvalidArgument, "couldnt extract metadata from request")
	}

	usernames, ok := md["username"]
	if !ok {
		return status.Error(codes.InvalidArgument, "couldnt extract username from request")
	}
	username := usernames[0]

	for _, u := range s.Users {
		if u.Name == username {
			return status.Error(codes.InvalidArgument, "username already taken")
		}
	}

	user := rooms.User{Name: username}
	s.Users[stream] = user
	defer delete(s.Users, stream)

	roomIDs, ok := md["room_id"]
	if !ok {
		return status.Error(codes.InvalidArgument, "couldnt extract room id from request")
	}
	roomID := roomIDs[0]

	allRooms, err := interactor.GetRooms()
	if err != nil {
		s.logger.Error(ctx, "couldnt fetch all rooms")
		return status.Error(codes.Internal, err.Error())
	}

	allRoomsIDs := make([]string, len(allRooms))
	for _, v := range allRooms {
		allRoomsIDs = append(allRoomsIDs, v.ID)
	}

	if !slices.Contains(allRoomsIDs, roomID) {
		return status.Error(codes.InvalidArgument, "no such room with given id")
	}

	if err := interactor.JoinRoom(roomID, user); err != nil {
		s.logger.Error(ctx, "couldnt join room", zap.String("room_id", roomID), zap.String("username", username))
		return status.Error(codes.Internal, err.Error())
	}
	defer func() {
		if err := interactor.LeaveRoom(roomID, user); err != nil {
			s.logger.Error(ctx, err.Error(), zap.String("room_id", roomID), zap.String("username", user.Name))
		}
	}()

	for {
		msg, err := stream.Recv()

		if err == io.EOF {
			return nil
		}

		if err != nil {
			s.logger.Error(ctx, err.Error())
			return status.Error(codes.Internal, err.Error())
		}

		method := msg.Method

		switch m := method.(type) {
		case *proto.RoomMethod_SendMessage:
			message := m.SendMessage
			text := message.Text

			roomUsers, err := interactor.GetRoomUsers(roomID)
			if err != nil {
				s.logger.Error(ctx, "couldnt fetch room users")
				return status.Error(codes.Internal, err.Error())
			}

			for stream, streamUser := range s.Users {
				if slices.Contains(roomUsers, streamUser) {
					userStream, ok := stream.(proto.RoomsService_JoinRoomServer)
					if !ok {
						s.logger.Error(ctx, "couldnt convert stream to JoinRoom server stream")
						return status.Error(codes.Internal, "couldnt process all room users")
					}

					message := &proto.MessageReceivedNotification{Text: text, Username: user.Name}
					method := &proto.RoomMethod{
						Method: &proto.RoomMethod_MessageReceived{
							MessageReceived: message,
						},
					}

					err := userStream.Send(method)
					if err != nil {
						s.logger.Error(ctx, "couldnt send message")
						return status.Error(codes.Internal, err.Error())
					}
				}
			}

		default:
			return status.Error(codes.InvalidArgument, "received invalid method")
		}
	}
}
