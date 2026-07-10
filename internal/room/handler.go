package room

import (
	"encoding/json"
	"net/http"

	"github.com/multiplayer-ludo/internal/dto"
	"github.com/multiplayer-ludo/internal/middleware"
)

type Handler struct {
	manager *Manager
}

func NewHandler(manager *Manager) *Handler {
	return &Handler{manager: manager}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.Handler) {
	mux.Handle("POST /api/rooms", authMiddleware(http.HandlerFunc(h.createRoom)))
	mux.Handle("POST /api/rooms/{code}/join", authMiddleware(http.HandlerFunc(h.joinRoom)))
	mux.Handle("GET /api/rooms/{code}", authMiddleware(http.HandlerFunc(h.getRoom)))
}

func (h *Handler) createRoom(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)

	room, err := h.manager.CreateRoom(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := dto.CreateRoomResponse{
		RoomCode: room.Code,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func (h *Handler) joinRoom(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	code := r.PathValue("code")

	if err := h.manager.JoinRoom(code, userID); err != nil {
		status := http.StatusInternalServerError
		if err == ErrRoomNotFound {
			status = http.StatusNotFound
		} else if err == ErrRoomFull || err == ErrRoomStarted {
			status = http.StatusBadRequest
		}
		http.Error(w, err.Error(), status)
		return
	}

	res := dto.JoinRoomResponse{
		RoomCode: code,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (h *Handler) getRoom(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")

	activeRoom, err := h.manager.GetRoomInfo(code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	activeRoom.RLock()
	defer activeRoom.RUnlock()

	var players []dto.RoomPlayerInfo
	for _, p := range activeRoom.Players {
		players = append(players, dto.RoomPlayerInfo{
			UserID:   p.UserID,
			Username: "", // In a full app, we'd fetch the username
			Color:    p.Color,
		})
	}

	res := dto.RoomInfo{
		Code:    activeRoom.Model.Code,
		HostID:  activeRoom.Model.HostID,
		Status:  activeRoom.Model.Status,
		Players: players,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
