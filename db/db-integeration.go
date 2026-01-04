package db

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func InitDB() {
	// Database initialization logic goes here
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL not set")
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatal(err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatal("DB ping failed:", err)
	}

	DB = pool
	log.Println("PostgreSQL connected ✅")
}
