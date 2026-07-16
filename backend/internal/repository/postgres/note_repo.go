package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"zametka/internal/domain"
)

type NoteRepository struct {
	pool *pgxpool.Pool
}

func NewNoteRepository(pool *pgxpool.Pool) *NoteRepository {
	return &NoteRepository{pool: pool}
}

func (r *NoteRepository) Create(ctx context.Context, n *domain.Note) error {
	if n.Reactions == nil {
		n.Reactions = []domain.Reaction{}
	}
	_, err := r.pool.Exec(ctx,
		`INSERT INTO notes (
			id, room_id, author_id, title, content, category, color, pinned, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		n.ID, n.RoomID, n.AuthorID, n.Title, n.Content, string(n.Category),
		n.Color, n.Pinned, n.CreatedAt, n.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert note: %w", err)
	}
	return nil
}

func (r *NoteRepository) GetByID(ctx context.Context, roomID, id string) (*domain.Note, error) {
	note, err := r.scanNote(ctx,
		`SELECT id, room_id, author_id, title, content, category, color, pinned, created_at, updated_at
		 FROM notes WHERE id = $1 AND room_id = $2`,
		id, roomID,
	)
	if err != nil {
		return nil, err
	}
	reactions, err := r.loadReactions(ctx, note.ID)
	if err != nil {
		return nil, err
	}
	note.Reactions = reactions
	return note, nil
}

func (r *NoteRepository) List(ctx context.Context, f domain.NoteFilter) ([]domain.Note, error) {
	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	args := []any{f.RoomID}
	var b strings.Builder
	b.WriteString(`SELECT id, room_id, author_id, title, content, category, color, pinned, created_at, updated_at
		FROM notes WHERE room_id = $1`)

	if f.Category != nil {
		args = append(args, string(*f.Category))
		fmt.Fprintf(&b, ` AND category = $%d`, len(args))
	}
	if f.Before != nil {
		args = append(args, *f.Before)
		fmt.Fprintf(&b, ` AND created_at < $%d`, len(args))
	}
	args = append(args, limit)
	fmt.Fprintf(&b, ` ORDER BY created_at DESC LIMIT $%d`, len(args))

	rows, err := r.pool.Query(ctx, b.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("list notes: %w", err)
	}
	defer rows.Close()

	notes := make([]domain.Note, 0)
	ids := make([]string, 0)
	for rows.Next() {
		var n domain.Note
		var category string
		if err := rows.Scan(
			&n.ID, &n.RoomID, &n.AuthorID, &n.Title, &n.Content, &category,
			&n.Color, &n.Pinned, &n.CreatedAt, &n.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan note: %w", err)
		}
		n.Category = domain.Category(category)
		n.Reactions = []domain.Reaction{}
		notes = append(notes, n)
		ids = append(ids, n.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate notes: %w", err)
	}

	byNote, err := r.loadReactionsForNotes(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range notes {
		if rs, ok := byNote[notes[i].ID]; ok {
			notes[i].Reactions = rs
		}
	}
	return notes, nil
}

func (r *NoteRepository) Update(ctx context.Context, roomID, id string, upd domain.NoteUpdate) (*domain.Note, error) {
	setParts := []string{"updated_at = $3"}
	args := []any{id, roomID, time.Now().UTC()}
	argN := 3

	if upd.Title != nil {
		argN++
		args = append(args, *upd.Title)
		setParts = append(setParts, fmt.Sprintf("title = $%d", argN))
	}
	if upd.Content != nil {
		argN++
		args = append(args, *upd.Content)
		setParts = append(setParts, fmt.Sprintf("content = $%d", argN))
	}
	if upd.Category != nil {
		argN++
		args = append(args, string(*upd.Category))
		setParts = append(setParts, fmt.Sprintf("category = $%d", argN))
	}
	if upd.Color != nil {
		argN++
		args = append(args, *upd.Color)
		setParts = append(setParts, fmt.Sprintf("color = $%d", argN))
	}
	if upd.Pinned != nil {
		argN++
		args = append(args, *upd.Pinned)
		setParts = append(setParts, fmt.Sprintf("pinned = $%d", argN))
	}

	query := fmt.Sprintf(
		`UPDATE notes SET %s WHERE id = $1 AND room_id = $2`,
		strings.Join(setParts, ", "),
	)
	tag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("update note: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, domain.ErrNoteNotFound
	}
	return r.GetByID(ctx, roomID, id)
}

func (r *NoteRepository) Delete(ctx context.Context, roomID, id string) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM notes WHERE id = $1 AND room_id = $2`,
		id, roomID,
	)
	if err != nil {
		return fmt.Errorf("delete note: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNoteNotFound
	}
	return nil
}

func (r *NoteRepository) UpsertReaction(ctx context.Context, roomID, noteID, memberID, emoji string) (*domain.Note, error) {
	tag, err := r.pool.Exec(ctx,
		`UPDATE notes SET updated_at = $3
		 WHERE id = $1 AND room_id = $2`,
		noteID, roomID, time.Now().UTC(),
	)
	if err != nil {
		return nil, fmt.Errorf("touch note: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, domain.ErrNoteNotFound
	}

	_, err = r.pool.Exec(ctx,
		`INSERT INTO note_reactions (note_id, member_id, emoji)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (note_id, member_id)
		 DO UPDATE SET emoji = EXCLUDED.emoji`,
		noteID, memberID, emoji,
	)
	if err != nil {
		return nil, fmt.Errorf("upsert reaction: %w", err)
	}
	return r.GetByID(ctx, roomID, noteID)
}

func (r *NoteRepository) RemoveReaction(ctx context.Context, roomID, noteID, memberID string) (*domain.Note, error) {
	tag, err := r.pool.Exec(ctx,
		`UPDATE notes SET updated_at = $3
		 WHERE id = $1 AND room_id = $2`,
		noteID, roomID, time.Now().UTC(),
	)
	if err != nil {
		return nil, fmt.Errorf("touch note: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, domain.ErrNoteNotFound
	}

	_, err = r.pool.Exec(ctx,
		`DELETE FROM note_reactions WHERE note_id = $1 AND member_id = $2`,
		noteID, memberID,
	)
	if err != nil {
		return nil, fmt.Errorf("remove reaction: %w", err)
	}
	return r.GetByID(ctx, roomID, noteID)
}

func (r *NoteRepository) scanNote(ctx context.Context, query string, args ...any) (*domain.Note, error) {
	var n domain.Note
	var category string
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&n.ID, &n.RoomID, &n.AuthorID, &n.Title, &n.Content, &category,
		&n.Color, &n.Pinned, &n.CreatedAt, &n.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNoteNotFound
		}
		return nil, fmt.Errorf("get note: %w", err)
	}
	n.Category = domain.Category(category)
	n.Reactions = []domain.Reaction{}
	return &n, nil
}

func (r *NoteRepository) loadReactions(ctx context.Context, noteID string) ([]domain.Reaction, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT member_id, emoji FROM note_reactions WHERE note_id = $1`,
		noteID,
	)
	if err != nil {
		return nil, fmt.Errorf("load reactions: %w", err)
	}
	defer rows.Close()

	out := make([]domain.Reaction, 0)
	for rows.Next() {
		var rr domain.Reaction
		if err := rows.Scan(&rr.MemberID, &rr.Emoji); err != nil {
			return nil, fmt.Errorf("scan reaction: %w", err)
		}
		out = append(out, rr)
	}
	return out, rows.Err()
}

func (r *NoteRepository) loadReactionsForNotes(ctx context.Context, noteIDs []string) (map[string][]domain.Reaction, error) {
	out := make(map[string][]domain.Reaction, len(noteIDs))
	if len(noteIDs) == 0 {
		return out, nil
	}

	rows, err := r.pool.Query(ctx,
		`SELECT note_id, member_id, emoji
		 FROM note_reactions
		 WHERE note_id = ANY($1)`,
		noteIDs,
	)
	if err != nil {
		return nil, fmt.Errorf("load reactions batch: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var noteID string
		var rr domain.Reaction
		if err := rows.Scan(&noteID, &rr.MemberID, &rr.Emoji); err != nil {
			return nil, fmt.Errorf("scan reaction batch: %w", err)
		}
		out[noteID] = append(out[noteID], rr)
	}
	return out, rows.Err()
}
