package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"go-gin-api-server/config"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func main() {
	var (
		action = flag.String("action", "", "Migration action: up, down, force, version")
		steps  = flag.Int("steps", 0, "Number of migration steps (for up/down)")
		test   = flag.Bool("test", false, "Use test database configuration")
	)
	flag.Parse()

	if *action == "" {
		fmt.Println("Usage: go run cmd/migrate/main.go -action=<up|down|force|version> [-steps=N] [-test]")
		fmt.Println("Examples:")
		fmt.Println("  go run cmd/migrate/main.go -action=up")
		fmt.Println("  go run cmd/migrate/main.go -action=down -steps=1")
		fmt.Println("  go run cmd/migrate/main.go -action=version")
		fmt.Println("  go run cmd/migrate/main.go -action=up -test")
		os.Exit(1)
	}

	// Load configuration
	var cfg *config.Config
	if *test {
		cfg = config.LoadTestConfig()
		fmt.Println("Using TEST database configuration")
	} else {
		cfg = config.LoadConfig()
		fmt.Println("Using PRODUCTION database configuration")
	}

	// Create database connection string
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	// Open database connection
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Create migrate instance
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("Failed to create postgres driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}

	// Execute migration action
	switch *action {
	case "up":
		if *steps > 0 {
			err = m.Steps(*steps)
		} else {
			err = m.Up()
		}
		if err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Failed to migrate up: %v", err)
		}
		fmt.Println("Migration up completed successfully")

	case "down":
		if *steps > 0 {
			err = m.Steps(-*steps)
		} else {
			err = m.Down()
		}
		if err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Failed to migrate down: %v", err)
		}
		fmt.Println("Migration down completed successfully")

	case "force":
		if *steps == 0 {
			log.Fatal("Force action requires -steps parameter")
		}
		err = m.Force(*steps)
		if err != nil {
			log.Fatalf("Failed to force migration: %v", err)
		}
		fmt.Printf("Migration forced to version %d\n", *steps)

	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			log.Fatalf("Failed to get migration version: %v", err)
		}
		fmt.Printf("Current migration version: %d (dirty: %t)\n", version, dirty)

	default:
		log.Fatalf("Unknown action: %s", *action)
	}
}
