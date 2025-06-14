package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var direction string
	var steps int
	flag.StringVar(&direction, "direction", "up", "Migration direction: up, down, force, or version")
	flag.IntVar(&steps, "steps", 0, "Number of migrations to apply (0 = all)")
	flag.Parse()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	m, err := migrate.New(
		"file://migrations",
		dbURL,
	)
	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}
	defer m.Close()

	switch direction {
	case "up":
		if steps > 0 {
			err = m.Steps(steps)
		} else {
			err = m.Up()
		}
		if err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration up failed: %v", err)
		}
		log.Println("Migration up completed successfully")

	case "down":
		if steps > 0 {
			err = m.Steps(-steps)
		} else {
			err = m.Down()
		}
		if err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration down failed: %v", err)
		}
		log.Println("Migration down completed successfully")

	case "force":
		version := steps
		if version == 0 {
			log.Fatal("--steps must specify the version number when using force")
		}
		if err := m.Force(version); err != nil {
			log.Fatalf("Force version failed: %v", err)
		}
		log.Printf("Forced migration version to %d", version)

	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			log.Fatalf("Failed to get version: %v", err)
		}
		fmt.Printf("Version: %d, Dirty: %v\n", version, dirty)

	default:
		log.Fatalf("Unknown direction: %s (use up, down, force, or version)", direction)
	}
}
