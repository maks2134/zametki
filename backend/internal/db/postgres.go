package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"zametka/internal/config"
)

const connectTimeout = 15 * time.Second

const schemaSQL = `
CREATE TABLE IF NOT EXISTS rooms (
    id TEXT PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS room_members (
    room_id TEXT NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    id TEXT NOT NULL,
    name TEXT NOT NULL,
    color TEXT NOT NULL,
    joined_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (room_id, id)
);

CREATE TABLE IF NOT EXISTS notes (
    id TEXT PRIMARY KEY,
    room_id TEXT NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    author_id TEXT NOT NULL,
    title TEXT NOT NULL DEFAULT '',
    content TEXT NOT NULL,
    category TEXT NOT NULL,
    color TEXT NOT NULL DEFAULT '',
    pinned BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS notes_room_created_idx
    ON notes (room_id, created_at DESC);

CREATE INDEX IF NOT EXISTS notes_room_category_created_idx
    ON notes (room_id, category, created_at DESC);

CREATE TABLE IF NOT EXISTS note_reactions (
    note_id TEXT NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    member_id TEXT NOT NULL,
    emoji TEXT NOT NULL,
    PRIMARY KEY (note_id, member_id)
);
`

func Connect(ctx context.Context, cfg config.Config) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()

	poolCfg, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	poolCfg.MaxConns = 8
	poolCfg.MinConns = 1

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("postgres connect: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres ping: %w", err)
	}

	return pool, nil
}

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, schemaSQL); err != nil {
		return fmt.Errorf("migrate schema: %w", err)
	}
	return nil
}
