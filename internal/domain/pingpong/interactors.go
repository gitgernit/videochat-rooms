package pingpong

type Interactor struct{}

func (p Interactor) Ping(ping Ping) (Pong, error) {
	counter := ping.Counter

	return Pong{Counter: counter + 1}, nil
}
