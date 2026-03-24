package sqlite

import (
	"database/sql"
	"fmt"
	"strings"
)

// migration holds a versioned set of SQL statements to be applied once.
type migration struct {
	version    int
	statements []string
}

var migrations = []migration{
	{
		version: 1,
		statements: []string{
			`CREATE TABLE IF NOT EXISTS schema_version (
    version INTEGER NOT NULL
)`,
			`CREATE TABLE IF NOT EXISTS day_plans (
    date        TEXT PRIMARY KEY,
    tasks       TEXT NOT NULL,
    goals       TEXT NOT NULL,
    score       INTEGER NOT NULL DEFAULT 0,
    created_at  TEXT NOT NULL
)`,
			`CREATE TABLE IF NOT EXISTS completion_records (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    date         TEXT NOT NULL,
    task_id      TEXT NOT NULL,
    status       TEXT NOT NULL,
    completed_at TEXT NOT NULL
)`,
			`CREATE TABLE IF NOT EXISTS streaks (
    id               INTEGER PRIMARY KEY CHECK (id = 1),
    current          INTEGER NOT NULL DEFAULT 0,
    longest          INTEGER NOT NULL DEFAULT 0,
    last_active_date TEXT NOT NULL
)`,
			`CREATE TABLE IF NOT EXISTS reflections (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    date       TEXT NOT NULL,
    text       TEXT NOT NULL,
    created_at TEXT NOT NULL
)`,
			`CREATE TABLE IF NOT EXISTS nudge_windows (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    start_time TEXT NOT NULL,
    end_time   TEXT NOT NULL,
    label      TEXT NOT NULL
)`,
		},
	},
}

// runMigrations applies any migrations whose version is greater than the
// currently recorded schema version. Migrations are applied in ascending
// version order inside a transaction and are never re-applied.
func runMigrations(db *sql.DB) error {
	// Ensure schema_version table exists so we can read/write it.
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_version (version INTEGER NOT NULL)`)
	if err != nil {
		return fmt.Errorf("create schema_version table: %w", err)
	}

	current, err := currentVersion(db)
	if err != nil {
		return fmt.Errorf("read schema version: %w", err)
	}

	for _, m := range migrations {
		if m.version <= current {
			continue
		}

		if err := applyMigration(db, m); err != nil {
			return err
		}
	}

	return nil
}

// applyMigration runs all statements for a single migration inside a
// transaction. The schema version is recorded as part of the same transaction
// so a partial migration can never leave the database in an inconsistent state.
func applyMigration(db *sql.DB, m migration) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin migration v%d: %w", m.version, err)
	}

	for _, stmt := range m.statements {
		if _, err := tx.Exec(strings.TrimSpace(stmt)); err != nil {
			tx.Rollback()
			return fmt.Errorf("apply migration v%d: %w", m.version, err)
		}
	}

	if _, err := tx.Exec(`INSERT INTO schema_version (version) VALUES (?)`, m.version); err != nil {
		tx.Rollback()
		return fmt.Errorf("record migration v%d: %w", m.version, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration v%d: %w", m.version, err)
	}

	return nil
}

func currentVersion(db *sql.DB) (int, error) {
	row := db.QueryRow(`SELECT version FROM schema_version ORDER BY version DESC LIMIT 1`)
	var v int
	if err := row.Scan(&v); err == sql.ErrNoRows {
		return 0, nil
	} else if err != nil {
		return 0, err
	}
	return v, nil
}
