package sqlite

import (
	"context"
	"database/sql"
	"errors"
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
	return withConn(ctx, db, func(conn *sql.Conn) error {
		statements := []string{
			`CREATE TABLE IF NOT EXISTS users (
                id TEXT PRIMARY KEY,
                email TEXT UNIQUE,
                password_hash TEXT,
                role TEXT NOT NULL,
                created_at TEXT NOT NULL,
                updated_at TEXT NOT NULL
            );`,
			`CREATE TABLE IF NOT EXISTS families (
                id TEXT PRIMARY KEY,
                name TEXT NOT NULL,
                owner_id TEXT,
                created_at TEXT NOT NULL
            );`,
			`CREATE TABLE IF NOT EXISTS family_members (
                id TEXT PRIMARY KEY,
                family_id TEXT NOT NULL,
                user_id TEXT,
                role TEXT NOT NULL,
                status TEXT NOT NULL DEFAULT 'ACTIVE',
                created_at TEXT NOT NULL,
                FOREIGN KEY(family_id) REFERENCES families(id)
            );`,
			`CREATE TABLE IF NOT EXISTS parents (
                id TEXT PRIMARY KEY
            );`,
			`CREATE TABLE IF NOT EXISTS children (
                id TEXT PRIMARY KEY,
                family_id TEXT,
                parent_id TEXT,
                name TEXT NOT NULL,
                xp INTEGER NOT NULL DEFAULT 0,
                gold INTEGER NOT NULL DEFAULT 0,
                created_at TEXT NOT NULL
            );`,
			`CREATE TABLE IF NOT EXISTS quests (
                id TEXT PRIMARY KEY,
                family_id TEXT,
                parent_id TEXT,
                title TEXT NOT NULL,
                description TEXT,
                xp INTEGER NOT NULL,
                gold INTEGER NOT NULL,
                created_at TEXT NOT NULL
            );`,
			`CREATE TABLE IF NOT EXISTS rewards (
                id TEXT PRIMARY KEY,
                family_id TEXT,
                parent_id TEXT,
                name TEXT NOT NULL,
                xp_threshold INTEGER NOT NULL,
                created_at TEXT NOT NULL
            );`,
			`CREATE TABLE IF NOT EXISTS assignments (
                id TEXT PRIMARY KEY,
                family_id TEXT,
                quest_id TEXT NOT NULL,
                child_id TEXT NOT NULL,
                status TEXT NOT NULL,
                created_at TEXT NOT NULL,
                completed_at TEXT,
                FOREIGN KEY(quest_id) REFERENCES quests(id),
                FOREIGN KEY(child_id) REFERENCES children(id)
            );`,
		}
		for _, stmt := range statements {
			if _, err := conn.ExecContext(ctx, stmt); err != nil {
				return fmt.Errorf("migrate execute: %w", err)
			}
		}

		if err := ensureColumn(ctx, conn, "children", "family_id", "ALTER TABLE children ADD COLUMN family_id TEXT"); err != nil {
			return err
		}
		if err := ensureColumn(ctx, conn, "quests", "family_id", "ALTER TABLE quests ADD COLUMN family_id TEXT"); err != nil {
			return err
		}
		if err := ensureColumn(ctx, conn, "rewards", "family_id", "ALTER TABLE rewards ADD COLUMN family_id TEXT"); err != nil {
			return err
		}
		if err := ensureColumn(ctx, conn, "assignments", "family_id", "ALTER TABLE assignments ADD COLUMN family_id TEXT"); err != nil {
			return err
		}

		if err := ensureIndex(ctx, conn, "idx_children_family", "CREATE INDEX IF NOT EXISTS idx_children_family ON children(family_id)"); err != nil {
			return err
		}
		if err := ensureIndex(ctx, conn, "idx_quests_family", "CREATE INDEX IF NOT EXISTS idx_quests_family ON quests(family_id)"); err != nil {
			return err
		}
		if err := ensureIndex(ctx, conn, "idx_rewards_family", "CREATE INDEX IF NOT EXISTS idx_rewards_family ON rewards(family_id)"); err != nil {
			return err
		}
		if err := ensureIndex(ctx, conn, "idx_assignments_family", "CREATE INDEX IF NOT EXISTS idx_assignments_family ON assignments(family_id)"); err != nil {
			return err
		}
		if err := ensureIndex(ctx, conn, "idx_families_owner", "CREATE INDEX IF NOT EXISTS idx_families_owner ON families(owner_id)"); err != nil {
			return err
		}
		if err := ensureIndex(ctx, conn, "idx_family_members_family", "CREATE INDEX IF NOT EXISTS idx_family_members_family ON family_members(family_id)"); err != nil {
			return err
		}
		if err := ensureIndex(ctx, conn, "idx_family_members_user", "CREATE INDEX IF NOT EXISTS idx_family_members_user ON family_members(user_id)"); err != nil {
			return err
		}

		updates := []string{
			`UPDATE children SET family_id = parent_id WHERE family_id IS NULL AND parent_id IS NOT NULL`,
			`UPDATE quests SET family_id = parent_id WHERE family_id IS NULL AND parent_id IS NOT NULL`,
			`UPDATE rewards SET family_id = parent_id WHERE family_id IS NULL AND parent_id IS NOT NULL`,
			`UPDATE assignments SET family_id = (
                SELECT family_id FROM children WHERE children.id = assignments.child_id
            ) WHERE family_id IS NULL`,
		}
		for _, stmt := range updates {
			if _, err := conn.ExecContext(ctx, stmt); err != nil {
				return fmt.Errorf("migrate update: %w", err)
			}
		}

		return nil
	})
}

func ensureColumn(ctx context.Context, conn *sql.Conn, table string, column string, addStmt string) error {
	query := fmt.Sprintf("PRAGMA table_info('%s')", table)
	rows, err := conn.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("pragma table_info %s: %w", table, err)
	}
	defer rows.Close()
	for rows.Next() {
		var (
			cid    int
			name   string
			ctype  string
			notnull int
			dflt   any
			pk     int
		)
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return err
		}
		if name == column {
			return nil
		}
	}
	if _, err := conn.ExecContext(ctx, addStmt); err != nil {
		return fmt.Errorf("add column %s.%s: %w", table, column, err)
	}
	return nil
}

func ensureIndex(ctx context.Context, conn *sql.Conn, name string, stmt string) error {
	var exists int
	if err := conn.QueryRowContext(ctx, `SELECT 1 FROM sqlite_master WHERE type = 'index' AND name = ?`, name).Scan(&exists); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
	}
	if exists == 1 {
		return nil
	}
	if _, err := conn.ExecContext(ctx, stmt); err != nil {
		return fmt.Errorf("create index %s: %w", name, err)
	}
	return nil
}
