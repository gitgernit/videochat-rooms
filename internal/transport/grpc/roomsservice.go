package grpc

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"io"
	"net/http"
	"slices"

	"github.com/gitgernit/videochat-contracts/proto/rooms/go/proto"
	"github.com/gitgernit/videochat-rooms/internal/domain/pingpong"
	"github.com/gitgernit/videochat-rooms/internal/domain/rooms"
	"github.com/gitgernit/videochat-rooms/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	usernameMetadata   = "username"
	roomNameMetadata   = "room_name"
	dispatcherUsername = "dispatcher"
)

func RoomsHeaderMatcher(key string) (string, bool) {
	switch key {
	case "Username":
		return usernameMetadata, true
	case "Room-Name":
		return roomNameMetadata, true
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

type RoomsService struct {
	proto.UnimplementedRoomsServiceServer
	logger               logger.Logger
	repository           rooms.Repository
	incomingRoomsChannel chan string
	Users                map[rooms.User]grpc.ServerStream
}

func NewRoomsService(logger logger.Logger, repository rooms.Repository, incomingRoomsChannel chan string) *RoomsService {
	return &RoomsService{
		logger:               logger,
		repository:           repository,
		Users:                make(map[rooms.User]grpc.ServerStream),
		incomingRoomsChannel: incomingRoomsChannel,
	}
}

func (s *RoomsService) PingPong(stream proto.RoomsService_PingPongServer) error {
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

func (s *RoomsService) ListenForRooms(in *proto.ListenForRoomsRequest, stream proto.RoomsService_ListenForRoomsServer) error {
	ctx := stream.Context()

	for {
		select {
		case roomName := <-s.incomingRoomsChannel:
			notification := &proto.NewRoomNotification{Name: roomName}
			if err := stream.Send(notification); err != nil {
				s.logger.Error(ctx, "could not send room notification", zap.String("room_id", roomName), zap.Error(err))
				return status.Error(codes.Internal, err.Error())
			}

		case <-ctx.Done():
			return nil
		}
	}
}

func (s *RoomsService) CreateRoom(ctx context.Context, req *proto.CreateRoomRequest) (*proto.CreateRoomResponse, error) {
	interactor := rooms.NewInteractor(s.logger, s.repository, s.incomingRoomsChannel)

	err := interactor.CreateRoom(req.Name)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &proto.CreateRoomResponse{Name: req.Name}, nil
}

func (s *RoomsService) JoinRoom(stream proto.RoomsService_JoinRoomServer) error {
	ctx := stream.Context()
	interactor := rooms.NewInteractor(s.logger, s.repository, s.incomingRoomsChannel)

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.InvalidArgument, "couldnt extract metadata from request")
	}

	usernames, ok := md["username"]
	if !ok {
		return status.Error(codes.InvalidArgument, "couldnt extract username from request")
	}
	username := usernames[0]

	user := rooms.User{Name: username, Id: uuid.New()}
	s.Users[user] = stream
	defer delete(s.Users, user)

	roomNames, ok := md["room_name"]
	if !ok {
		return status.Error(codes.InvalidArgument, "couldnt extract room name from request")
	}
	roomName := roomNames[0]

	roomUsers, err := interactor.GetRoomUsers(roomName)
	if err != nil {
		return status.Error(codes.Internal, "couldnt fetch room users")
	}

	for _, u := range roomUsers {
		if u.Name == username {
			return status.Error(codes.InvalidArgument, "username already taken")
		}
	}

	allRooms, err := interactor.GetRooms()
	if err != nil {
		s.logger.Error(ctx, "couldnt fetch all rooms")
		return status.Error(codes.Internal, err.Error())
	}

	allRoomsIDs := make([]string, len(allRooms))
	for _, v := range allRooms {
		allRoomsIDs = append(allRoomsIDs, v.Name)
	}

	if !slices.Contains(allRoomsIDs, roomName) {
		return status.Error(codes.InvalidArgument, "no such room with given id")
	}

	if err := interactor.JoinRoom(roomName, user); err != nil {
		s.logger.Error(ctx, "couldnt join room", zap.String("room_id", roomName), zap.String("username", username))
		return status.Error(codes.Internal, err.Error())
	}

	err = s.sendRoomUsers(ctx, interactor, roomName)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	defer func(s *RoomsService, ctx context.Context, interactor rooms.Interactor, roomID string) {
		err := s.sendRoomUsers(ctx, interactor, roomID)
		if err != nil {
			s.logger.Error(ctx, "couldnt send room users upon user leaving room")
		}
	}(s, ctx, interactor, roomName)
	defer func(interactor rooms.Interactor, id string, user rooms.User) {
		err := interactor.LeaveRoom(id, user)
		if err != nil {
			s.logger.Error(ctx, err.Error(), zap.String("room_id", roomName), zap.String("username", user.Name))
		}
	}(interactor, roomName, user)

	for {
		msg, err := stream.Recv()

		if err == io.EOF {
			return nil
		}

		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		method := msg.Method

		switch m := method.(type) {
		case *proto.RoomMethod_SendMessage:
			message := m.SendMessage
			text := message.Text

			roomUsers, err := interactor.GetRoomUsers(roomName)
			if err != nil {
				s.logger.Error(ctx, "couldnt fetch room users")
				return status.Error(codes.Internal, err.Error())
			}

			for streamUser, stream := range s.Users {
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

		case *proto.RoomMethod_SendSdp:
			message := m.SendSdp
			sdps := message.Sdp

			roomUsers, err := interactor.GetRoomUsers(roomName)
			if err != nil {
				s.logger.Error(ctx, "couldnt fetch room users")
				return status.Error(codes.Internal, err.Error())
			}

			for _, sdp := range sdps {
				user := roomUsers[0]

				for _, roomUser := range roomUsers {
					if roomUser.Name == sdp.Username {
						user = roomUser
						break
					}
				}

				stream := s.Users[user]

				userStream, ok := stream.(proto.RoomsService_JoinRoomServer)
				if !ok {
					s.logger.Error(ctx, "couldnt convert stream to JoinRoom server stream")
					return status.Error(codes.Internal, "couldnt process all room users")
				}

				method := &proto.RoomMethod{
					Method: &proto.RoomMethod_SdpReceived{
						SdpReceived: &proto.SDPReceivedNotification{
							Type: sdp.Type,
							Sdp:  sdp.Sdp,
							To:   sdp.Username,
							From: username,
						},
					},
				}

				err := userStream.Send(method)
				if err != nil {
					s.logger.Error(ctx, "couldnt send sdp")
					return status.Error(codes.Internal, err.Error())
				}
			}

		default:
			return status.Error(codes.InvalidArgument, "received invalid method")
		}
	}
}

func (s *RoomsService) sendRoomUsers(ctx context.Context, interactor rooms.Interactor, roomID string) error {
	roomUsers, err := interactor.GetRoomUsers(roomID)
	if err != nil {
		return status.Error(codes.Internal, "couldnt fetch room users")
	}

	protoRoomUsers := make([]*proto.User, len(roomUsers))
	for i, u := range roomUsers {
		user := proto.User{Id: u.Id.String(), Username: u.Name}
		protoRoomUsers[i] = &user
	}

	method := &proto.RoomMethod{
		Method: &proto.RoomMethod_RoomUsers_{
			RoomUsers_: &proto.RoomUsers{
				Users: protoRoomUsers,
			},
		},
	}

	for streamUser, stream := range s.Users {
		if slices.Contains(roomUsers, streamUser) {
			userStream, ok := stream.(proto.RoomsService_JoinRoomServer)
			if !ok {
				s.logger.Error(ctx, "couldnt convert stream to JoinRoom server stream")
				return fmt.Errorf("couldnt process all room users")
			}

			err = userStream.Send(method)
			if err != nil {
				return fmt.Errorf("couldnt send room users")
			}
		}
	}

	return nil
}
