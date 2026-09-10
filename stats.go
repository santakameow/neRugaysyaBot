package main

import (
	"database/sql"
	"errors"
	"fmt"

	_ "modernc.org/sqlite"
)

// function for initialisation database with path
// that comes from .env
func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            user_id INTEGER PRIMARY KEY,
            swear_count INTEGER NOT NULL DEFAULT 0
        );
    `)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("create users table: %w", err)
	}

	// CREATE TABLE IF NOT EXISTS wont fix a pre existing table with the
	// wrong schema (eg a dev/test file). Fail fast with a clear message
	// instead of cryptic "no such column" errors at query time.
	rows, err := db.Query(`PRAGMA table_info(users)`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("inspect users table: %w", err)
	}
	defer rows.Close()

	cols := map[string]bool{}
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			db.Close()
			return nil, fmt.Errorf("inspect users table: %w", err)
		}
		cols[name] = true
	}
	if err := rows.Err(); err != nil {
		db.Close()
		return nil, fmt.Errorf("inspect users table: %w", err)
	}
	if !cols["user_id"] || !cols["swear_count"] {
		db.Close()
		return nil, fmt.Errorf("table users has unexpected schema (have columns %v, want user_id and swear_count); move the stale db file aside and restart", cols)
	}

	return db, nil
}

// function to increment user count of swear in database
func incrementSwearCount(db *sql.DB, userID int64) error {
	_, err := db.Exec(`
        INSERT INTO users (user_id, swear_count)
        VALUES (?, 1)
        ON CONFLICT(user_id)
        DO UPDATE SET swear_count = swear_count + 1
    `, userID)

	return err
}

// function to get swear count, returns int
// can be useful somewhere, i dont know
// unknown users have count 0 (no row yet), not an error
func getSwearCount(db *sql.DB, userID int64) (int, error) {
	var count int

	err := db.QueryRow(
		"SELECT swear_count FROM users WHERE user_id = ?",
		userID,
	).Scan(&count)

	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	return count, nil
}
