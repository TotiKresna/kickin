package config

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func ConnectDB(cfg *Config) *sql.DB {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=verify-ca&sslrootcert=%s",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.SSLRootCert,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping DB: %v", err)
	}

	return db
}
