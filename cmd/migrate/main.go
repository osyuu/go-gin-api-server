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

func runMigration(action string, steps int, test bool) error {
	// Load configuration
	var cfg *config.Config
	if test {
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
		return fmt.Errorf("failed to open database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	// Create migrate instance
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %v", err)
	}

	// Execute migration action
	switch action {
	case "up":
		if steps > 0 {
			err = m.Steps(steps)
		} else {
			err = m.Up()
		}
		if err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("failed to migrate up: %v", err)
		}
		fmt.Println("Migration up completed successfully")

	case "down":
		if steps > 0 {
			err = m.Steps(-steps)
		} else {
			err = m.Down()
		}
		if err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("failed to migrate down: %v", err)
		}
		fmt.Println("Migration down completed successfully")

	case "force":
		if steps == 0 {
			return fmt.Errorf("force action requires -steps parameter")
		}
		err = m.Force(steps)
		if err != nil {
			return fmt.Errorf("failed to force migration: %v", err)
		}
		fmt.Printf("Migration forced to version %d\n", steps)

	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			return fmt.Errorf("failed to get migration version: %v", err)
		}
		fmt.Printf("Current migration version: %d (dirty: %t)\n", version, dirty)

	default:
		return fmt.Errorf("unknown action: %s", action)
	}

	return nil
}

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

	// Run migration and handle errors
	if err := runMigration(*action, *steps, *test); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
}
