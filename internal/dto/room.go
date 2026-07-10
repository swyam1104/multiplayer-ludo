package dto

import "github.com/multiplayer-ludo/internal/models"

type CreateRoomResponse struct {
	RoomCode string `json:"room_code"`
}

type JoinRoomResponse struct {
	RoomCode string `json:"room_code"`
	// In a full implementation, we'd return the current players, state, etc.
}

type RoomInfo struct {
	Code    string              `json:"code"`
	HostID  string              `json:"host_id"`
	Status  models.RoomStatus   `json:"status"`
	Players []RoomPlayerInfo    `json:"players"`
}

type RoomPlayerInfo struct {
	UserID   string            `json:"user_id"`
	Username string            `json:"username"`
	Color    models.TokenColor `json:"color"`
}
