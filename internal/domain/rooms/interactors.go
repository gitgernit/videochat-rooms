package rooms

import (
	"github.com/google/uuid"
	"gitlab.crja72.ru/gospec/go5/rooms/pkg/logger"
)

type Interactor struct {
	logger     logger.Logger
	repository Repository
}

func NewInteractor(logger logger.Logger, repository Repository) Interactor {
	return Interactor{
		logger:     logger,
		repository: repository,
	}
}

func (i Interactor) CreateRoom() (string, error) {
	id := uuid.New()

	err := i.repository.CreateRoom(id.String())
	if err != nil {
		return "", err
	}

	return id.String(), nil
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
