package pingpong

type PingInteractor struct{}

func (p PingInteractor) Ping(ping Ping) (Pong, error) {
	counter := ping.Counter

	return Pong{Counter: counter + 1}, nil
}
