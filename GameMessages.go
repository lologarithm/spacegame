package main

type GameMessage interface {
}

type GameMessageValues struct {
	FromUser uint32
	Client   *Client
}

type LoginMessage struct {
	GameMessageValues
	LoggingIn bool
}

type SetThrustMessage struct {
	GameMessageValues
	ThrustPercent []uint8
}

type PhysicsUpdateMessage struct {
	GameMessageValues
	Ships []*Ship
}
