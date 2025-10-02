package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Open opens (or creates) a SQLite database at the given path and applies pragmatic defaults
// suitable for LiteFS/Litestream usage. The caller must ensure close is called when finished.
func Open(ctx context.Context, path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil && !os.IsExist(err) {
		return nil, fmt.Errorf("create sqlite directory: %w", err)
	}
	dsn := fmt.Sprintf("file:%s?_busy_timeout=5000&_fk=1", path)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	// SQLite is not great with many concurrent writers. Limit to a single writer connection.
	db.SetConnMaxLifetime(0)
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	// Apply pragmas on first connection.
	if err := withConn(ctx, db, func(conn *sql.Conn) error {
		pragmas := []string{
			"PRAGMA journal_mode=WAL",
			"PRAGMA synchronous=NORMAL",
			"PRAGMA foreign_keys=ON",
		}
		for _, p := range pragmas {
			if _, err := conn.ExecContext(ctx, p); err != nil {
				return fmt.Errorf("apply pragma %q: %w", p, err)
			}
		}
		return nil
	}); err != nil {
		db.Close()
		return nil, err
	}

	if err := migrate(ctx, db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func withConn(ctx context.Context, db *sql.DB, fn func(*sql.Conn) error) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	return fn(conn)
}

func migrate(ctx context.Context, db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS parents (
            id TEXT PRIMARY KEY
        );`,
		`CREATE TABLE IF NOT EXISTS children (
            id TEXT PRIMARY KEY,
            parent_id TEXT NOT NULL,
            name TEXT NOT NULL,
            xp INTEGER NOT NULL DEFAULT 0,
            gold INTEGER NOT NULL DEFAULT 0,
            created_at TEXT NOT NULL,
            FOREIGN KEY(parent_id) REFERENCES parents(id)
        );`,
		`CREATE INDEX IF NOT EXISTS idx_children_parent ON children(parent_id);`,
		`CREATE TABLE IF NOT EXISTS quests (
            id TEXT PRIMARY KEY,
            parent_id TEXT NOT NULL,
            title TEXT NOT NULL,
            description TEXT,
            xp INTEGER NOT NULL,
            gold INTEGER NOT NULL,
            created_at TEXT NOT NULL,
            FOREIGN KEY(parent_id) REFERENCES parents(id)
        );`,
		`CREATE INDEX IF NOT EXISTS idx_quests_parent ON quests(parent_id);`,
		`CREATE TABLE IF NOT EXISTS rewards (
            id TEXT PRIMARY KEY,
            parent_id TEXT NOT NULL,
            name TEXT NOT NULL,
            xp_threshold INTEGER NOT NULL,
            created_at TEXT NOT NULL,
            FOREIGN KEY(parent_id) REFERENCES parents(id)
        );`,
		`CREATE INDEX IF NOT EXISTS idx_rewards_parent ON rewards(parent_id);`,
		`CREATE TABLE IF NOT EXISTS assignments (
            id TEXT PRIMARY KEY,
            quest_id TEXT NOT NULL,
            child_id TEXT NOT NULL,
            status TEXT NOT NULL,
            created_at TEXT NOT NULL,
            completed_at TEXT,
            FOREIGN KEY(quest_id) REFERENCES quests(id),
            FOREIGN KEY(child_id) REFERENCES children(id)
        );`,
		`CREATE INDEX IF NOT EXISTS idx_assignments_child ON assignments(child_id);`,
		`CREATE INDEX IF NOT EXISTS idx_assignments_quest ON assignments(quest_id);`,
	}

	return withConn(ctx, db, func(conn *sql.Conn) error {
		for _, stmt := range stmts {
			if _, err := conn.ExecContext(ctx, stmt); err != nil {
				return fmt.Errorf("migrate execute: %w", err)
			}
		}
		return nil
	})
}
