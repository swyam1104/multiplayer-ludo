package websocket

import (
	"fmt"
	"log"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/multiplayer-ludo/internal/config"
	"github.com/multiplayer-ludo/internal/room"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// allow all origins for now
		return true
	},
}

type Handler struct {
	hub         *Hub
	config      *config.Config
	roomManager *room.Manager
}

func NewHandler(hub *Hub, cfg *config.Config, rm *room.Manager) *Handler {
	return &Handler{
		hub:         hub,
		config:      cfg,
		roomManager: rm,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /ws", h.serveWs)
}

func (h *Handler) serveWs(w http.ResponseWriter, r *http.Request) {
	roomCode := r.URL.Query().Get("room")
	tokenStr := r.URL.Query().Get("token")

	if roomCode == "" || tokenStr == "" {
		http.Error(w, "missing room or token", http.StatusBadRequest)
		return
	}

	// Validate JWT
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(h.config.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "invalid claims", http.StatusUnauthorized)
		return
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		http.Error(w, "invalid user id", http.StatusUnauthorized)
		return
	}

	// Verify the user is actually in this room
	activeRoom, err := h.roomManager.GetRoomInfo(roomCode)
	if err != nil {
		http.Error(w, "room not found", http.StatusNotFound)
		return
	}

	activeRoom.RLock()
	var inRoom bool
	for _, p := range activeRoom.Players {
		if p.UserID == userID {
			inRoom = true
			break
		}
	}
	activeRoom.RUnlock()

	if !inRoom {
		http.Error(w, "user not in room", http.StatusForbidden)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{
		Hub:      h.hub,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		UserID:   userID,
		RoomCode: roomCode,
	}

	client.Hub.Register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.WritePump()
	go client.ReadPump()
}
