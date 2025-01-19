package db

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

var PostgresDB *sql.DB

func InitPostgres(dsn string) {
	var err error
	PostgresDB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	if err := PostgresDB.Ping(); err != nil {
		log.Fatalf("PostgreSQL not reachable: %v", err)
	}

	log.Println("Connected to PostgreSQL!")
}
