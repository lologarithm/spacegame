package main

type GameMessage interface {
}

type GameMessageValues struct {
	FromUser int32
	Client   *Client
}

type LoginMessage struct {
	GameMessageValues
	LoggingIn bool
}

type SetThrustMessage struct {
	GameMessageValues
	ThrustPercent []int16
}
