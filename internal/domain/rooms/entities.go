package rooms

type User struct {
	Name string
}

type Room struct {
	ID    string
	Users []User
}
