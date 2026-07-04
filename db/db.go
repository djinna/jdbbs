package db

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"sort"
	"strconv"
	"strings"

	_ "modernc.org/sqlite"
)

//go:generate go tool github.com/sqlc-dev/sqlc/cmd/sqlc generate

//go:embed migrations/*.sql
var migrationFS embed.FS

// Open opens an sqlite database and prepares pragmas suitable for a small web app.
func Open(path string) (*sql.DB, error) {
	// Apply pragmas via the DSN so EVERY pooled connection gets them (a bare
	// `PRAGMA` via db.Exec only affects one connection). foreign_keys and
	// busy_timeout are per-connection in SQLite.
	sep := "?"
	if strings.Contains(path, "?") {
		sep = "&"
	}
	dsn := path + sep + "_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	// SQLite allows a single writer; serialize all access through one connection
	// so concurrent writes queue on busy_timeout instead of failing SQLITE_BUSY,
	// and so the per-connection pragmas above always apply.
	db.SetMaxOpenConns(1)
	return db, nil
}

// RunMigrations executes database migrations in numeric order (NNN-*.sql),
// similar in spirit to exed's exedb.RunMigrations.
func RunMigrations(db *sql.DB) error {
	entries, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}
	var migrations []string
	pat := regexp.MustCompile(`^(\d{3})-.*\.sql$`)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if pat.MatchString(name) {
			migrations = append(migrations, name)
		}
	}
	sort.Strings(migrations)

	executed := make(map[int]bool)
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='migrations'").Scan(&tableName)
	switch {
	case err == nil:
		rows, err := db.Query("SELECT migration_number FROM migrations")
		if err != nil {
			return fmt.Errorf("query executed migrations: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var n int
			if err := rows.Scan(&n); err != nil {
				return fmt.Errorf("scan migration number: %w", err)
			}
			executed[n] = true
		}
	case errors.Is(err, sql.ErrNoRows):
		slog.Info("db: migrations table not found; running all migrations")
	default:
		return fmt.Errorf("check migrations table: %w", err)
	}

	for _, m := range migrations {
		match := pat.FindStringSubmatch(m)
		if len(match) != 2 {
			return fmt.Errorf("invalid migration filename: %s", m)
		}
		n, err := strconv.Atoi(match[1])
		if err != nil {
			return fmt.Errorf("parse migration number %s: %w", m, err)
		}
		if executed[n] {
			continue
		}
		if err := executeMigration(db, m, n); err != nil {
			return fmt.Errorf("execute %s: %w", m, err)
		}
		slog.Info("db: applied migration", "file", m, "number", n)
	}
	return nil
}

func executeMigration(db *sql.DB, filename string, number int) error {
	content, err := migrationFS.ReadFile("migrations/" + filename)
	if err != nil {
		return fmt.Errorf("read %s: %w", filename, err)
	}
	// Run each migration in its own transaction so a mid-file failure rolls back
	// cleanly and is never recorded as applied. Without this, a partial apply
	// (e.g. a bare ALTER that half-ran) makes the re-run fail and bricks startup.
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin %s: %w", filename, err)
	}
	defer tx.Rollback()
	if _, err := tx.Exec(string(content)); err != nil {
		return fmt.Errorf("exec %s: %w", filename, err)
	}
	// Record from the runner too (idempotent on the PK), so a file that omits its
	// own INSERT is still tracked, and recording commits atomically with the DDL.
	if _, err := tx.Exec(
		`INSERT OR IGNORE INTO migrations (migration_number, migration_name) VALUES (?, ?)`,
		number, filename,
	); err != nil {
		return fmt.Errorf("record %s: %w", filename, err)
	}
	return tx.Commit()
}
