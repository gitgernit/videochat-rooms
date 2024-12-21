package rooms

import (
	"github.com/google/uuid"
	"gitlab.crja72.ru/gospec/go5/rooms/pkg/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

type Interactor struct {
	logger          logger.Logger
	repository      Repository
	newRoomsChannel chan string
}

func NewInteractor(logger logger.Logger, repository Repository, newRoomsChannel chan string) Interactor {
	return Interactor{
		logger:          logger,
		repository:      repository,
		newRoomsChannel: newRoomsChannel,
	}
}

func (i Interactor) CreateRoom() (string, error) {
	id := uuid.New()

	err := i.repository.CreateRoom(id.String())
	if err != nil {
		return "", err
	}

	select {
	case i.newRoomsChannel <- id.String():
		return id.String(), nil
	case <-time.After(5 * time.Second):
		return "", status.Error(codes.Unavailable, "couldnt send the room to a coordinator")
	}
}

func (i Interactor) JoinRoom(id string, user User) error {
	err := i.repository.JoinRoom(id, user)
	return err
}

func (i Interactor) LeaveRoom(id string, user User) error {
	err := i.repository.LeaveRoom(id, user)
	return err
}

func (i Interactor) GetRoomUsers(id string) ([]User, error) {
	users, err := i.repository.GetRoomUsers(id)
	return users, err
}

func (i Interactor) GetRooms() ([]Room, error) {
	rooms, err := i.repository.GetRooms()
	return rooms, err
}
