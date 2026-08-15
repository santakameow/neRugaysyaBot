package main

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func statistics() {
	db, err := sql.Open("sqlite", "./stats.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	fmt.Println(db.Stats())

	db.Query("select * from users")
}

func main() {
	statistics()
}
