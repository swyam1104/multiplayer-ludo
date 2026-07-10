package websocket

import "encoding/json"

type EventType string

const (
	EventJoinRoom          EventType = "join_room"
	EventLeaveRoom         EventType = "leave_room"
	EventStartGame         EventType = "start_game"
	EventRollDice          EventType = "roll_dice"
	EventMoveToken         EventType = "move_token"
	EventPing              EventType = "ping"
	EventPong              EventType = "pong"

	// Outgoing events
	EventRoomUpdate        EventType = "room_update"
	EventGameStarted       EventType = "game_started"
	EventDiceRolled        EventType = "dice_rolled"
	EventValidMoves        EventType = "valid_moves"
	EventTokenMoved        EventType = "token_moved"
	EventTurnChanged       EventType = "turn_changed"
	EventPlayerDisconnected EventType = "player_disconnected"
	EventGameOver          EventType = "game_over"
	EventError             EventType = "error"
)

type Event struct {
	Type    EventType       `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
