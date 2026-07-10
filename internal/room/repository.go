package room

import (
	"database/sql"

	"github.com/multiplayer-ludo/internal/models"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateTables() error {
	roomsQuery := `
	CREATE TABLE IF NOT EXISTS rooms (
		id VARCHAR(36) PRIMARY KEY,
		code VARCHAR(6) UNIQUE NOT NULL,
		host_id VARCHAR(36) NOT NULL,
		status ENUM('waiting', 'playing', 'finished') DEFAULT 'waiting',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (host_id) REFERENCES users(id)
	);
	`
	if _, err := r.db.Exec(roomsQuery); err != nil {
		return err
	}

	playersQuery := `
	CREATE TABLE IF NOT EXISTS room_players (
		room_id VARCHAR(36) NOT NULL,
		user_id VARCHAR(36) NOT NULL,
		color ENUM('red', 'green', 'yellow', 'blue') NOT NULL,
		joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (room_id, user_id),
		FOREIGN KEY (room_id) REFERENCES rooms(id),
		FOREIGN KEY (user_id) REFERENCES users(id)
	);
	`
	_, err := r.db.Exec(playersQuery)
	return err
}

func (r *Repository) CreateRoom(room *models.Room) error {
	query := `INSERT INTO rooms (id, code, host_id, status, created_at) VALUES (?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query, room.ID, room.Code, room.HostID, room.Status, room.CreatedAt)
	return err
}

func (r *Repository) AddPlayer(player *models.RoomPlayer) error {
	query := `INSERT INTO room_players (room_id, user_id, color, joined_at) VALUES (?, ?, ?, ?)`
	_, err := r.db.Exec(query, player.RoomID, player.UserID, player.Color, player.JoinedAt)
	return err
}

func (r *Repository) GetRoomByCode(code string) (*models.Room, error) {
	query := `SELECT id, code, host_id, status, created_at FROM rooms WHERE code = ?`
	row := r.db.QueryRow(query, code)

	var room models.Room
	err := row.Scan(&room.ID, &room.Code, &room.HostID, &room.Status, &room.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, err
	}
	return &room, nil
}
