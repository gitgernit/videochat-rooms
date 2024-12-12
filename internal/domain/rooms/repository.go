package rooms

type Repository interface {
	CreateRoom(id string) error
	JoinRoom(id string, user User) error
	LeaveRoom(id string, user User) error
	GetRoomUsers(id string) ([]User, error)
	GetRooms() ([]Room, error)
}
