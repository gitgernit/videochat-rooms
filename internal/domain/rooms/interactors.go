package rooms

import (
	"github.com/gitgernit/videochat-rooms/pkg/logger"
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

func (i Interactor) CreateRoom(name string) error {
	err := i.repository.CreateRoom(name)
	if err != nil {
		return err
	}

	select {
	case i.newRoomsChannel <- name:
		return nil
	case <-time.After(5 * time.Second):
		return status.Error(codes.Unavailable, "couldnt send the room to a coordinator")
	}
}

func (i Interactor) JoinRoom(name string, user User) error {
	err := i.repository.JoinRoom(name, user)
	return err
}

func (i Interactor) LeaveRoom(name string, user User) error {
	err := i.repository.LeaveRoom(name, user)
	return err
}

func (i Interactor) GetRoomUsers(name string) ([]User, error) {
	users, err := i.repository.GetRoomUsers(name)
	return users, err
}

func (i Interactor) GetRooms() ([]Room, error) {
	rooms, err := i.repository.GetRooms()
	return rooms, err
}
