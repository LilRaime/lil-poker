package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const startingChips = 1000

type User struct {
	UUID            string `json:"uuid"`
	Username        string `json:"username"`
	Password        string `json:"-"`
	Chips           int    `json:"chips"`
	RebuysRemaining int    `json:"rebuys_remaining"`
	IsGuest         bool   `json:"is_guest"`
}

func queryUser(db *sql.DB, query, arg string) (*User, error) {
	var u User
	err := db.QueryRow(query, arg).Scan(&u.UUID, &u.Username, &u.Password, &u.Chips, &u.RebuysRemaining, &u.IsGuest)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &u, nil
}

func GetUserByUsername(db *sql.DB, username string) (*User, error) {
	return queryUser(db, `SELECT uuid, username, password, chips, rebuys_remaining, is_guest FROM users WHERE username = $1`, username)
}

func GetUserByUUID(db *sql.DB, uuidStr string) (*User, error) {
	return queryUser(db, `SELECT uuid, username, password, chips, rebuys_remaining, is_guest FROM users WHERE uuid = $1`, uuidStr)
}

func createUser(db *sql.DB, username, passwordHash string, isGuest bool) (*User, error) {
	uuidStr := uuid.New().String()
	query := `
	INSERT INTO users (uuid, username, password, chips, rebuys_remaining, is_guest)
	VALUES ($1, $2, $3, $4, 3, $5)
	RETURNING uuid, username, password, chips, rebuys_remaining, is_guest`

	var u User
	err := db.QueryRow(query, uuidStr, username, passwordHash, startingChips, isGuest).Scan(
		&u.UUID, &u.Username, &u.Password, &u.Chips, &u.RebuysRemaining, &u.IsGuest,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert user: %w", err)
	}
	return &u, nil
}

func CreateUser(db *sql.DB, username, passwordHash string) (*User, error) {
	return createUser(db, username, passwordHash, false)
}

func CreateGuestUser(db *sql.DB, username, passwordHash string) (*User, error) {
	return createUser(db, username, passwordHash, true)
}

func UpdateUserChipsAndRebuys(db *sql.DB, uuidStr string, chips, rebuys int) error {
	query := `UPDATE users SET chips = $1, rebuys_remaining = $2 WHERE uuid = $3`
	_, err := db.Exec(query, chips, rebuys, uuidStr)
	if err != nil {
		return fmt.Errorf("failed to update user chips and rebuys: %w", err)
	}
	return nil
}

func UpdateUserChips(db *sql.DB, uuidStr string, chips int) error {
	query := `UPDATE users SET chips = $1 WHERE uuid = $2`
	_, err := db.Exec(query, chips, uuidStr)
	if err != nil {
		return fmt.Errorf("failed to update user chips: %w", err)
	}
	return nil
}

func UpdateUsersChipsTx(db *sql.DB, updates map[string]int) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`UPDATE users SET chips = $1 WHERE uuid = $2`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for u, chips := range updates {
		_, err := stmt.Exec(chips, u)
		if err != nil {
			return fmt.Errorf("failed to execute update for user %s: %w", u, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func DeleteStaleGuests(db *sql.DB, olderThan time.Duration) (int64, error) {
	res, err := db.Exec(
		`DELETE FROM users WHERE is_guest = TRUE AND created_at < $1`,
		time.Now().Add(-olderThan),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to delete stale guests: %w", err)
	}
	n, _ := res.RowsAffected()
	return n, nil
}
