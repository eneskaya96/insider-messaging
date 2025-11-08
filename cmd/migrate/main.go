package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/eneskaya/insider-messaging/pkg/config"
	_ "github.com/lib/pq"
)

func main() {
	log.Println("Starting database migration...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := sql.Open("postgres", cfg.Database.DSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("Connected to database successfully")

	migrationPath := "migrations/001_create_messages_table.sql"
	content, err := os.ReadFile(migrationPath)
	if err != nil {
		log.Fatalf("Failed to read migration file: %v", err)
	}

	log.Printf("Executing migration: %s", filepath.Base(migrationPath))

	if _, err := db.Exec(string(content)); err != nil {
		log.Fatalf("Failed to execute migration: %v", err)
	}

	log.Println("Migration completed successfully!")
}
