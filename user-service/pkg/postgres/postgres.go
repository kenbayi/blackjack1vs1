package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

type Config struct {
	Host         string `env:"POSTGRES_HOST" envDefault:"localhost"`
	Port         string `env:"POSTGRES_PORT" envDefault:"5432"`
	Database     string `env:"POSTGRES_DB" envDefault:"user_service_db"`
	Username     string `env:"POSTGRES_USER" envDefault:"postgres"`
	Password     string `env:"POSTGRES_PASSWORD" envDefault:""`
	SSLMode      string `env:"POSTGRES_SSL_MODE" envDefault:"disable"`
	MaxOpenConns int    `env:"POSTGRES_MAX_OPEN_CONNS" envDefault:"10"`
	MaxIdleConns int    `env:"POSTGRES_MAX_IDLE_CONNS" envDefault:"5"`
}

type DB struct {
	Conn   *sql.DB
	client *sql.DB
}

// NewDB creates connection to PostgreSQL and returns the DB struct.
func NewDB(ctx context.Context, cfg Config) (*DB, error) {
	connStr := cfg.genConnectURL()

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("connection to PostgreSQL Error: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Minute * 5)

	// Verify connection
	err = db.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("ping connection PostgreSQL Error: %w", err)
	}

	pgDB := &DB{
		Conn:   db,
		client: db,
	}

	go pgDB.reconnectOnFailure(ctx)

	return pgDB, nil
}

// reconnectOnFailure implements db reconnection if ping was unsuccessful.
func (db *DB) reconnectOnFailure(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)

	for {
		select {
		case <-ticker.C:
			err := db.client.PingContext(ctx)
			if err != nil {
				log.Printf("lost connection to PostgreSQL: %v", err)

				// Attempt to reconnect
				err = db.client.PingContext(ctx)
				if err == nil {
					log.Println("reconnected to PostgreSQL successfully")
				}
			}
		case <-ctx.Done():
			ticker.Stop()
			err := db.client.Close()
			if err != nil {
				log.Printf("PostgreSQL close connection error: %v", err)
				return
			}

			log.Println("PostgreSQL connection is closed successfully")
			return
		}
	}
}

func (db *DB) Ping(ctx context.Context) error {
	err := db.client.PingContext(ctx)
	if err != nil {
		return fmt.Errorf("PostgreSQL connection error: %w", err)
	}
	return nil
}
