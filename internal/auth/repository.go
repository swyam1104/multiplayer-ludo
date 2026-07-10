package auth

import (
	"database/sql"
	"errors"

	"github.com/multiplayer-ludo/internal/models"
)

var ErrUserNotFound = errors.New("user not found")
var ErrEmailTaken = errors.New("email already taken")
var ErrUsernameTaken = errors.New("username already taken")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id VARCHAR(36) PRIMARY KEY,
		username VARCHAR(255) UNIQUE NOT NULL,
		email VARCHAR(255) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		wins INT DEFAULT 0,
		losses INT DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := r.db.Exec(query)
	return err
}

func (r *Repository) CreateUser(u *models.User) error {
	query := `INSERT INTO users (id, username, email, password_hash, created_at) VALUES (?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query, u.ID, u.Username, u.Email, u.PasswordHash, u.CreatedAt)
	if err != nil {
		// Basic error parsing, should ideally check for specific mysql error codes
		return err
	}
	return nil
}

func (r *Repository) GetUserByEmail(email string) (*models.User, error) {
	query := `SELECT id, username, email, password_hash, wins, losses, created_at FROM users WHERE email = ?`
	row := r.db.QueryRow(query, email)

	var u models.User
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.Wins, &u.Losses, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *Repository) GetUserByID(id string) (*models.User, error) {
	query := `SELECT id, username, email, password_hash, wins, losses, created_at FROM users WHERE id = ?`
	row := r.db.QueryRow(query, id)

	var u models.User
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.Wins, &u.Losses, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}
