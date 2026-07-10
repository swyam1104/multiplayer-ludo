package websocket

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/multiplayer-ludo/internal/game"
	"github.com/multiplayer-ludo/internal/models"
	"github.com/multiplayer-ludo/internal/room"
)

type Hub struct {
	// Registered clients mapped by RoomCode then UserID
	Rooms map[string]map[string]*Client

	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan EventMessage

	roomManager *room.Manager
	mu          sync.RWMutex
}

type EventMessage struct {
	RoomCode string
	Event    Event
}

func NewHub(rm *room.Manager) *Hub {
	return &Hub{
		Rooms:       make(map[string]map[string]*Client),
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		Broadcast:   make(chan EventMessage),
		roomManager: rm,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			if h.Rooms[client.RoomCode] == nil {
				h.Rooms[client.RoomCode] = make(map[string]*Client)
			}
			h.Rooms[client.RoomCode][client.UserID] = client
			h.mu.Unlock()

			// Broadcast room update
			go h.notifyRoomUpdate(client.RoomCode)

		case client := <-h.Unregister:
			h.mu.Lock()
			if roomClients, ok := h.Rooms[client.RoomCode]; ok {
				if _, ok := roomClients[client.UserID]; ok {
					delete(roomClients, client.UserID)
					close(client.Send)
					if len(roomClients) == 0 {
						delete(h.Rooms, client.RoomCode)
					}
				}
			}
			h.mu.Unlock()

			// Broadcast disconnect
			disconnectEvent := Event{
				Type: EventPlayerDisconnected,
				Payload: toJSON(map[string]string{
					"player_id": client.UserID,
				}),
			}
			h.BroadcastToRoom(client.RoomCode, disconnectEvent)
			go h.notifyRoomUpdate(client.RoomCode)

		case msg := <-h.Broadcast:
			h.BroadcastToRoom(msg.RoomCode, msg.Event)
		}
	}
}

func (h *Hub) BroadcastToRoom(roomCode string, event Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	roomClients, ok := h.Rooms[roomCode]
	if !ok {
		return
	}

	eventBytes, err := json.Marshal(event)
	if err != nil {
		log.Printf("error marshaling event: %v", err)
		return
	}

	for _, client := range roomClients {
		select {
		case client.Send <- eventBytes:
		default:
			// If buffer is full, we assume the client is dead
			close(client.Send)
			delete(roomClients, client.UserID)
		}
	}
}

func (h *Hub) ProcessEvent(client *Client, ev Event) {
	activeRoom, err := h.roomManager.GetRoomInfo(client.RoomCode)
	if err != nil {
		return
	}

	activeRoom.Lock()
	defer activeRoom.Unlock()
	engine := activeRoom.Engine

	switch ev.Type {
	case EventStartGame:
		// Only host can start
		if activeRoom.Model.HostID != client.UserID {
			return
		}
		if err := engine.StartGame(); err != nil {
			h.sendError(client, "start_game_failed", err.Error())
			return
		}
		activeRoom.Model.Status = models.RoomStatusPlaying

		// Broadcast game started
		h.BroadcastToRoom(client.RoomCode, Event{Type: EventGameStarted})
		
		// Send valid moves to current player if they happen to not need to roll (e.g. if we add a different mode). For Ludo, you always roll first.
		h.BroadcastToRoom(client.RoomCode, Event{
			Type: EventTurnChanged,
			Payload: toJSON(map[string]string{
				"player_id": engine.GetPlayerByColor(engine.CurrentTurnColor).UserID,
			}),
		})
		go h.notifyRoomUpdate(client.RoomCode)

	case EventRollDice:
		val, err := engine.RollDice(client.UserID)
		if err != nil {
			h.sendError(client, "roll_failed", err.Error())
			return
		}

		h.BroadcastToRoom(client.RoomCode, Event{
			Type: EventDiceRolled,
			Payload: toJSON(map[string]interface{}{
				"player_id": client.UserID,
				"value":     val,
			}),
		})

		// Next player if 0
		if engine.DiceValue == 0 {
			h.BroadcastToRoom(client.RoomCode, Event{
				Type: EventTurnChanged,
				Payload: toJSON(map[string]string{
					"player_id": engine.GetPlayerByColor(engine.CurrentTurnColor).UserID,
				}),
			})
		} else {
			// Provide valid moves only to the roller
			validMoves := engine.GetValidMoves(engine.CurrentTurnColor, engine.DiceValue)
			movesEv := Event{
				Type: EventValidMoves,
				Payload: toJSON(map[string]interface{}{
					"token_ids": validMoves,
				}),
			}
			client.Send <- toJSON(movesEv)
		}

	case EventMoveToken:
		var payload struct {
			TokenID int `json:"token_id"`
		}
		if err := json.Unmarshal(ev.Payload, &payload); err != nil {
			return
		}

		err := engine.MoveToken(client.UserID, payload.TokenID)
		if err != nil {
			h.sendError(client, "move_failed", err.Error())
			return
		}

		// Broadly notify of the move
		h.BroadcastToRoom(client.RoomCode, Event{
			Type: EventTokenMoved,
			Payload: toJSON(map[string]interface{}{
				"player_id": client.UserID,
				"token_id":  payload.TokenID,
			}),
		})

		if engine.State == game.StateFinished {
			h.BroadcastToRoom(client.RoomCode, Event{
				Type: EventGameOver,
				Payload: toJSON(map[string]interface{}{
					"winner_id": engine.GetPlayerByColor(engine.Winner).UserID,
				}),
			})
			activeRoom.Model.Status = models.RoomStatusFinished
			go h.notifyRoomUpdate(client.RoomCode)
			return
		}

		// Notify next turn
		h.BroadcastToRoom(client.RoomCode, Event{
			Type: EventTurnChanged,
			Payload: toJSON(map[string]string{
				"player_id": engine.GetPlayerByColor(engine.CurrentTurnColor).UserID,
			}),
		})
	}
}

func (h *Hub) sendError(client *Client, code, message string) {
	errEv := Event{
		Type: EventError,
		Payload: toJSON(ErrorPayload{
			Code:    code,
			Message: message,
		}),
	}
	client.Send <- toJSON(errEv)
}

func (h *Hub) notifyRoomUpdate(roomCode string) {
	// Fetch room from manager and broadcast
	activeRoom, err := h.roomManager.GetRoomInfo(roomCode)
	if err != nil {
		return
	}
	
	activeRoom.RLock()
	// Build a simple view of the room to broadcast
	var players []string
	for _, p := range activeRoom.Players {
		players = append(players, p.UserID)
	}
	status := string(activeRoom.Model.Status)
	activeRoom.RUnlock()

	ev := Event{
		Type: EventRoomUpdate,
		Payload: toJSON(map[string]interface{}{
			"status":  status,
			"players": players,
		}),
	}
	
	h.BroadcastToRoom(roomCode, ev)
}

func toJSON(v interface{}) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}
