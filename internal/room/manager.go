package room

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/multiplayer-ludo/internal/game"
	"github.com/multiplayer-ludo/internal/models"
)

var (
	ErrRoomNotFound = errors.New("room not found")
	ErrRoomFull     = errors.New("room is full")
	ErrRoomStarted  = errors.New("room already started")
)

// ActiveRoom holds the live state of a room in memory.
// It will eventually hold a pointer to the game engine state.
type ActiveRoom struct {
	sync.RWMutex
	Model   *models.Room
	Players []*models.RoomPlayer
	Engine  *game.Engine
}

type Manager struct {
	repo   *Repository
	rooms  map[string]*ActiveRoom // mapped by Room Code
	mu     sync.RWMutex
}

func NewManager(repo *Repository) *Manager {
	return &Manager{
		repo:  repo,
		rooms: make(map[string]*ActiveRoom),
	}
}

func generateRoomCode() string {
	bytes := make([]byte, 3)
	if _, err := rand.Read(bytes); err != nil {
		return "123456" // fallback
	}
	return hex.EncodeToString(bytes)
}

func (m *Manager) CreateRoom(hostID string) (*models.Room, error) {
	code := generateRoomCode()
	room := &models.Room{
		ID:        uuid.NewString(),
		Code:      code,
		HostID:    hostID,
		Status:    models.RoomStatusWaiting,
		CreatedAt: time.Now().UTC(),
	}

	if err := m.repo.CreateRoom(room); err != nil {
		return nil, err
	}

	hostPlayer := &models.RoomPlayer{
		RoomID:   room.ID,
		UserID:   hostID,
		Color:    models.ColorRed, // default to Red for host
		JoinedAt: time.Now().UTC(),
	}

	if err := m.repo.AddPlayer(hostPlayer); err != nil {
		return nil, err
	}

	m.mu.Lock()
	m.rooms[code] = &ActiveRoom{
		Model:   room,
		Players: []*models.RoomPlayer{hostPlayer},
		Engine:  game.NewEngine(),
	}
	m.mu.Unlock()

	// Host is added to engine
	m.rooms[code].Engine.AddPlayer(hostID, models.ColorRed)

	return room, nil
}

func (m *Manager) JoinRoom(code, userID string) error {
	m.mu.RLock()
	activeRoom, exists := m.rooms[code]
	m.mu.RUnlock()

	if !exists {
		// Attempt to load from DB in a full implementation,
		// but for MVP memory is source of truth for active games.
		return ErrRoomNotFound
	}

	activeRoom.Lock()
	defer activeRoom.Unlock()

	if activeRoom.Model.Status != models.RoomStatusWaiting {
		return ErrRoomStarted
	}

	if len(activeRoom.Players) >= 4 {
		return ErrRoomFull
	}

	// Check if already in room
	for _, p := range activeRoom.Players {
		if p.UserID == userID {
			return nil // Already joined
		}
	}

	// Assign color
	usedColors := make(map[models.TokenColor]bool)
	for _, p := range activeRoom.Players {
		usedColors[p.Color] = true
	}

	availableColors := []models.TokenColor{models.ColorRed, models.ColorGreen, models.ColorYellow, models.ColorBlue}
	var assignedColor models.TokenColor
	for _, c := range availableColors {
		if !usedColors[c] {
			assignedColor = c
			break
		}
	}

	player := &models.RoomPlayer{
		RoomID:   activeRoom.Model.ID,
		UserID:   userID,
		Color:    assignedColor,
		JoinedAt: time.Now().UTC(),
	}

	if err := m.repo.AddPlayer(player); err != nil {
		return err
	}

	activeRoom.Players = append(activeRoom.Players, player)
	activeRoom.Engine.AddPlayer(userID, assignedColor)

	// Note: We will broadcast a WebSocket event here eventually

	return nil
}

func (m *Manager) GetRoomInfo(code string) (*ActiveRoom, error) {
	m.mu.RLock()
	activeRoom, exists := m.rooms[code]
	m.mu.RUnlock()

	if !exists {
		return nil, ErrRoomNotFound
	}

	return activeRoom, nil
}
