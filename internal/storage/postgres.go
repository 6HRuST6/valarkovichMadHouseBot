package storage

import (
	"database/sql"
	"errors"
)

type Storage struct {
	DB *sql.DB
}

type User struct {
	ID           int64
	TelegramID   int64
	Username     string
	FirstName    string
	LastName     string
	RegisteredAt string
}

func New(db *sql.DB) *Storage {
	return &Storage{DB: db}
}

func (s *Storage) Init() error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id BIGSERIAL PRIMARY KEY,
		telegram_id BIGINT NOT NULL UNIQUE,
		username TEXT,
		first_name TEXT NOT NULL,
		last_name TEXT,
		registered_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	`
	_, err := s.DB.Exec(query)
	return err
}

func (s *Storage) RegisterUser(telegramID int64, username, firstName, lastName string) error {

	if telegramID == 0 {
		return errors.New("telegram id is empty")
	}

	_, err := s.DB.Exec(`
		INSERT INTO users (telegram_id, username, first_name, last_name)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (telegram_id)
		DO UPDATE SET
			username = EXCLUDED.username,
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name
	`, telegramID, username, firstName, lastName)

	return err

}

func (s *Storage) GetAllUsers() ([]User, error) {

	rows, err := s.DB.Query(
		`
		SELECT
			id,
			telegram_id,
			COALESCE(username, ''),
			first_name,
			COALESCE(last_name, ''),
			registered_at::text
		FROM users
		ORDER BY id
	`)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users []User

	for rows.Next() {
		var u User
		if err := rows.Scan(
			&u.ID,
			&u.TelegramID,
			&u.Username,
			&u.FirstName,
			&u.LastName,
			&u.RegisteredAt,
		); err != nil {
			return nil, err
		}

		users = append(users, u)
	}

	return users, rows.Err()

}

func (s *Storage) GetUserByTelegramID(telegramID int64) (User, error) {
	var u User

	err := s.DB.QueryRow(`
		SELECT
			id,
			telegram_id,
			COALESCE(username, ''),
			first_name,
			COALESCE(last_name, ''),
			registered_at::text
		FROM users
		WHERE telegram_id = $1
	`, telegramID).Scan(
		&u.ID,
		&u.TelegramID,
		&u.Username,
		&u.FirstName,
		&u.LastName,
		&u.RegisteredAt,
	)

	return u, err
}

func (s *Storage) DeleteUserByTelegramID(telegramID int64) error {
	_, err := s.DB.Exec(`
		DELETE FROM users
		WHERE telegram_id = $1
	`, telegramID)

	return err
}
