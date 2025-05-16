package migrations

import (
	"database/sql"
	"log"
)

func RunMigrations(db *sql.DB) {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		role TEXT NOT NULL
	);
	`

	_, err := db.Exec(query)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("Database migrated successfully.")
}
