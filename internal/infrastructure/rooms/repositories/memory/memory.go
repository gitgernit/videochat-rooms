package memory

import (
	"fmt"
	"slices"

	"github.com/gitgernit/videochat-rooms/internal/domain/rooms"
)

type Repository struct {
	rooms map[string]rooms.Room
}

func NewRepository() *Repository {
	return &Repository{
		rooms: make(map[string]rooms.Room),
	}
}

func (r *Repository) CreateRoom(name string) error {
	room := rooms.Room{Name: name, Users: make([]rooms.User, 0)}
	r.rooms[room.Name] = room

	return nil
}

func (r *Repository) JoinRoom(name string, user rooms.User) error {
	room, ok := r.rooms[name]
	if !ok {
		return fmt.Errorf("no such room with given name")
	}

	room.Users = append(room.Users, user)
	r.rooms[name] = room

	return nil
}

func (r *Repository) LeaveRoom(id string, user rooms.User) error {
	room, ok := r.rooms[id]
	if !ok {
		return fmt.Errorf("no such room with given id")
	}

	users := room.Users
	if !slices.Contains(users, user) {
		return fmt.Errorf("no such user in room")
	}

	index := 0

	for i, v := range users {
		if v == user {
			index = i
			break
		}
	}

	users[index] = users[len(users)-1]
	users = users[:len(users)-1]

	room.Users = users
	r.rooms[id] = room

	return nil
}

func (r *Repository) GetRoomUsers(id string) ([]rooms.User, error) {
	room, ok := r.rooms[id]
	if !ok {
		return nil, fmt.Errorf("no such room with given id")
	}

	users := room.Users
	return users, nil
}

func (r *Repository) GetRooms() ([]rooms.Room, error) {
	roomsValues := make([]rooms.Room, len(r.rooms))

	for _, v := range r.rooms {
		roomsValues = append(roomsValues, v)
	}

	return roomsValues, nil
}
