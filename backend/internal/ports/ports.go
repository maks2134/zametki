package ports

import (
	"context"

	"zametka/internal/domain"
)

type RoomRepository interface {
	Create(ctx context.Context, room *domain.Room) error
	GetByID(ctx context.Context, id string) (*domain.Room, error)
	GetByCode(ctx context.Context, code string) (*domain.Room, error)
	AddMember(ctx context.Context, code string, m domain.Member) (*domain.Room, error)
}

type NoteRepository interface {
	Create(ctx context.Context, n *domain.Note) error
	GetByID(ctx context.Context, roomID, id string) (*domain.Note, error)
	List(ctx context.Context, f domain.NoteFilter) ([]domain.Note, error)
	Update(ctx context.Context, roomID, id string, upd domain.NoteUpdate) (*domain.Note, error)
	Delete(ctx context.Context, roomID, id string) error
	UpsertReaction(ctx context.Context, roomID, noteID, memberID, emoji string) (*domain.Note, error)
	RemoveReaction(ctx context.Context, roomID, noteID, memberID string) (*domain.Note, error)
}

type RoomService interface {
	Create(ctx context.Context, title, name, color string) (room *domain.Room, token string, err error)
	Join(ctx context.Context, code, name, color string) (room *domain.Room, token string, err error)
	Get(ctx context.Context, roomID string) (*domain.Room, error)
}

type NoteService interface {
	List(ctx context.Context, roomID string, f domain.NoteFilter) ([]domain.Note, error)
	Create(ctx context.Context, roomID, authorID string, in domain.NoteCreate) (*domain.Note, error)
	Update(ctx context.Context, roomID, memberID, noteID string, upd domain.NoteUpdate) (*domain.Note, error)
	Delete(ctx context.Context, roomID, memberID, noteID string) error
	AddReaction(ctx context.Context, roomID, memberID, noteID, emoji string) (*domain.Note, error)
	RemoveReaction(ctx context.Context, roomID, memberID, noteID string) (*domain.Note, error)
}

type Broadcaster interface {
	Broadcast(roomID string, ev Event)
}

type Event struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

type Claims struct {
	RoomID   string
	MemberID string
}

type TokenIssuer interface {
	Issue(roomID, memberID string) (string, error)
	Parse(token string) (Claims, error)
}
