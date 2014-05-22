package server

type GameMessage interface {
}

type GameMessageValues struct {
	FromUser uint32
	Client   *Client
}

// 1. Player creates new game.
// 2. Player joins game.
// 3. Player leaves game.
type PlayerGameMessage struct {
	GameMessageValues
	GameAction byte
	GameId     uint32
}

// 1. Game changes state 2. Game removed(could be a state)
type GameStatusMessage struct {
	GameMessageValues
	GameStatus byte
	GameId     uint32
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
