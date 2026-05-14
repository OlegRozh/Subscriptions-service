package migrations

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func Up(databaseURL, migrationsDir string) error {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			fmt.Printf("failed to close database: %v\n", err)
		}
	}(db)

	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	return nil
}
