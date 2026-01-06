package helper

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool
var ctx = context.Background()

func InitDB() {
	// Database initialization logic goes here
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL not set")
	}

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatal(err)
	}

	// Disable prepared statement caching to avoid SQLSTATE 42P05 errors
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatal(err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatal("DB ping failed:", err)
	}

	DB = pool
	log.Println("PostgreSQL connected ✅")
}
