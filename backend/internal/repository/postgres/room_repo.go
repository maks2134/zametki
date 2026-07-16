package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"zametka/internal/domain"
)

type RoomRepository struct {
	pool *pgxpool.Pool
}

func NewRoomRepository(pool *pgxpool.Pool) *RoomRepository {
	return &RoomRepository{pool: pool}
}

func (r *RoomRepository) Create(ctx context.Context, room *domain.Room) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin create room: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO rooms (id, code, title, created_at) VALUES ($1, $2, $3, $4)`,
		room.ID, room.Code, room.Title, room.CreatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrCodeTaken
		}
		return fmt.Errorf("insert room: %w", err)
	}

	for _, m := range room.Members {
		_, err = tx.Exec(ctx,
			`INSERT INTO room_members (room_id, id, name, color, joined_at)
			 VALUES ($1, $2, $3, $4, $5)`,
			room.ID, m.ID, m.Name, m.Color, m.JoinedAt,
		)
		if err != nil {
			return fmt.Errorf("insert member: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit create room: %w", err)
	}
	return nil
}

func (r *RoomRepository) GetByID(ctx context.Context, id string) (*domain.Room, error) {
	return r.getRoom(ctx, `WHERE r.id = $1`, id)
}

func (r *RoomRepository) GetByCode(ctx context.Context, code string) (*domain.Room, error) {
	return r.getRoom(ctx, `WHERE r.code = $1`, code)
}

func (r *RoomRepository) AddMember(ctx context.Context, code string, m domain.Member) (*domain.Room, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin add member: %w", err)
	}
	defer tx.Rollback(ctx)

	var roomID string
	err = tx.QueryRow(ctx,
		`SELECT id FROM rooms WHERE code = $1 FOR UPDATE`,
		code,
	).Scan(&roomID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrRoomNotFound
		}
		return nil, fmt.Errorf("lock room: %w", err)
	}

	var count int
	if err := tx.QueryRow(ctx,
		`SELECT COUNT(*) FROM room_members WHERE room_id = $1`,
		roomID,
	).Scan(&count); err != nil {
		return nil, fmt.Errorf("count members: %w", err)
	}
	if count >= 2 {
		return nil, domain.ErrRoomFull
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO room_members (room_id, id, name, color, joined_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		roomID, m.ID, m.Name, m.Color, m.JoinedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert member: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit add member: %w", err)
	}

	return r.GetByID(ctx, roomID)
}

func (r *RoomRepository) getRoom(ctx context.Context, where string, arg any) (*domain.Room, error) {
	var room domain.Room
	err := r.pool.QueryRow(ctx,
		`SELECT r.id, r.code, r.title, r.created_at
		 FROM rooms r `+where,
		arg,
	).Scan(&room.ID, &room.Code, &room.Title, &room.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrRoomNotFound
		}
		return nil, fmt.Errorf("get room: %w", err)
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, name, color, joined_at
		 FROM room_members
		 WHERE room_id = $1
		 ORDER BY joined_at ASC`,
		room.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}
	defer rows.Close()

	room.Members = make([]domain.Member, 0, 2)
	for rows.Next() {
		var m domain.Member
		if err := rows.Scan(&m.ID, &m.Name, &m.Color, &m.JoinedAt); err != nil {
			return nil, fmt.Errorf("scan member: %w", err)
		}
		room.Members = append(room.Members, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate members: %w", err)
	}
	return &room, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
