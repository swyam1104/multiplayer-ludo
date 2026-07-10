package models

import "time"

type RoomStatus string

const (
	RoomStatusWaiting  RoomStatus = "waiting"
	RoomStatusPlaying  RoomStatus = "playing"
	RoomStatusFinished RoomStatus = "finished"
)

type Room struct {
	ID        string     `json:"id"`
	Code      string     `json:"code"`
	HostID    string     `json:"host_id"`
	Status    RoomStatus `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
}

type TokenColor string

const (
	ColorRed    TokenColor = "red"
	ColorGreen  TokenColor = "green"
	ColorYellow TokenColor = "yellow"
	ColorBlue   TokenColor = "blue"
)

type RoomPlayer struct {
	RoomID   string     `json:"room_id"`
	UserID   string     `json:"user_id"`
	Color    TokenColor `json:"color"`
	JoinedAt time.Time  `json:"joined_at"`
}
