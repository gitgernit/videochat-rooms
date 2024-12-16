package rooms

import "github.com/google/uuid"

type User struct {
	Id   uuid.UUID
	Name string
}

type Room struct {
	ID    string
	Users []User
}
