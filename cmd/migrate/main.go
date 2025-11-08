package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"

	"github.com/eneskaya/insider-messaging/pkg/config"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func main() {
	var (
		migrationsPath = flag.String("path", "migrations", "Path to migration files")
		command        = flag.String("cmd", "up", "Migration command: up, down, version, force")
		steps          = flag.Int("steps", -1, "Number of migrations to run (for down command)")
		version        = flag.Int("version", -1, "Force version (for force command)")
	)
	flag.Parse()

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

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("Failed to create migration driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", *migrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}

	switch *command {
	case "up":
		log.Println("Running migrations up...")
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration up failed: %v", err)
		}
		log.Println("Migrations up completed successfully!")

	case "down":
		log.Println("Running migrations down...")
		if *steps > 0 {
			if err := m.Steps(-*steps); err != nil && err != migrate.ErrNoChange {
				log.Fatalf("Migration down failed: %v", err)
			}
		} else {
			if err := m.Down(); err != nil && err != migrate.ErrNoChange {
				log.Fatalf("Migration down failed: %v", err)
			}
		}
		log.Println("Migrations down completed successfully!")

	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			log.Fatalf("Failed to get version: %v", err)
		}
		log.Printf("Current version: %d (dirty: %v)\n", version, dirty)

	case "force":
		if *version < 0 {
			log.Fatal("Version flag required for force command")
		}
		log.Printf("Forcing version to %d...\n", *version)
		if err := m.Force(*version); err != nil {
			log.Fatalf("Force version failed: %v", err)
		}
		log.Println("Version forced successfully!")

	default:
		log.Fatalf("Unknown command: %s. Use: up, down, version, or force", *command)
	}
}
