package rooms

type Repository interface {
	CreateRoom(name string) error
	JoinRoom(name string, user User) error
	LeaveRoom(name string, user User) error
	GetRoomUsers(name string) ([]User, error)
	GetRooms() ([]Room, error)
}
