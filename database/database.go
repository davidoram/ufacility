package database

import (
	"database/sql"
	"log"
)

func MigrateDatabase(db *sql.DB) error {
	log.Println("Running db migrations...")
	err := db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	return nil
}
