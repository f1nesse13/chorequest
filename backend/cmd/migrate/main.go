package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/joho/godotenv"

	sqlitedb "chorequest/backend/internal/sqlite"
)

func main() {
	ctx := context.Background()
	_ = godotenv.Load()

	var sqlitePath string
	flag.StringVar(&sqlitePath, "sqlite", os.Getenv("SQLITE_PATH"), "Path to SQLite database")
	flag.Parse()

	if sqlitePath == "" {
		sqlitePath = "./data/dev.sqlite"
	}

	db, err := sqlitedb.Open(ctx, sqlitePath)
	if err != nil {
		log.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	log.Printf("SQLite schema ready at %s", sqlitePath)
}
