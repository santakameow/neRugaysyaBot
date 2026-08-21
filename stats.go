package main

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

func InitDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "/data/stats.db")
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            user_id INTEGER PRIMARY KEY,
            swear_count INTEGER NOT NULL DEFAULT 0
        );
    `)
	if err != nil {
		panic(err)
	}

	return db, nil
}

func incrementSwearCount(db *sql.DB, userID int64) error {
	_, err := db.Exec(`
        INSERT INTO users (user_id, swear_count)
        VALUES (?, 1)
        ON CONFLICT(user_id)
        DO UPDATE SET swear_count = swear_count + 1
    `, userID)

	return err
}

func getSwearCount(db *sql.DB, userID int64) (int, error) {
	var count int

	err := db.QueryRow(
		"SELECT swear_count FROM users WHERE user_id = ?",
		userID,
	).Scan(&count)

	if err != nil {
		return 0, err
	}

	return count, nil
}
