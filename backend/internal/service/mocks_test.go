package service

import (
	"context"
	"sync"

	"zametka/internal/domain"
	"zametka/internal/ports"
)

type mockRoomRepo struct {
	createFn    func(ctx context.Context, room *domain.Room) error
	getByIDFn   func(ctx context.Context, id string) (*domain.Room, error)
	getByCodeFn func(ctx context.Context, code string) (*domain.Room, error)
	addMemberFn func(ctx context.Context, code string, m domain.Member) (*domain.Room, error)
}

func (m *mockRoomRepo) Create(ctx context.Context, room *domain.Room) error {
	if m.createFn != nil {
		return m.createFn(ctx, room)
	}
	return nil
}

func (m *mockRoomRepo) GetByID(ctx context.Context, id string) (*domain.Room, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, domain.ErrRoomNotFound
}

func (m *mockRoomRepo) GetByCode(ctx context.Context, code string) (*domain.Room, error) {
	if m.getByCodeFn != nil {
		return m.getByCodeFn(ctx, code)
	}
	return nil, domain.ErrRoomNotFound
}

func (m *mockRoomRepo) AddMember(ctx context.Context, code string, member domain.Member) (*domain.Room, error) {
	if m.addMemberFn != nil {
		return m.addMemberFn(ctx, code, member)
	}
	return nil, domain.ErrRoomNotFound
}

type mockNoteRepo struct {
	createFn         func(ctx context.Context, n *domain.Note) error
	getByIDFn        func(ctx context.Context, roomID, id string) (*domain.Note, error)
	listFn           func(ctx context.Context, f domain.NoteFilter) ([]domain.Note, error)
	updateFn         func(ctx context.Context, roomID, id string, upd domain.NoteUpdate) (*domain.Note, error)
	deleteFn         func(ctx context.Context, roomID, id string) error
	upsertReactionFn func(ctx context.Context, roomID, noteID, memberID, emoji string) (*domain.Note, error)
	removeReactionFn func(ctx context.Context, roomID, noteID, memberID string) (*domain.Note, error)
}

func (m *mockNoteRepo) Create(ctx context.Context, n *domain.Note) error {
	if m.createFn != nil {
		return m.createFn(ctx, n)
	}
	return nil
}

func (m *mockNoteRepo) GetByID(ctx context.Context, roomID, id string) (*domain.Note, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, roomID, id)
	}
	return nil, domain.ErrNoteNotFound
}

func (m *mockNoteRepo) List(ctx context.Context, f domain.NoteFilter) ([]domain.Note, error) {
	if m.listFn != nil {
		return m.listFn(ctx, f)
	}
	return nil, nil
}

func (m *mockNoteRepo) Update(ctx context.Context, roomID, id string, upd domain.NoteUpdate) (*domain.Note, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, roomID, id, upd)
	}
	return nil, domain.ErrNoteNotFound
}

func (m *mockNoteRepo) Delete(ctx context.Context, roomID, id string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, roomID, id)
	}
	return domain.ErrNoteNotFound
}

func (m *mockNoteRepo) UpsertReaction(ctx context.Context, roomID, noteID, memberID, emoji string) (*domain.Note, error) {
	if m.upsertReactionFn != nil {
		return m.upsertReactionFn(ctx, roomID, noteID, memberID, emoji)
	}
	return nil, domain.ErrNoteNotFound
}

func (m *mockNoteRepo) RemoveReaction(ctx context.Context, roomID, noteID, memberID string) (*domain.Note, error) {
	if m.removeReactionFn != nil {
		return m.removeReactionFn(ctx, roomID, noteID, memberID)
	}
	return nil, domain.ErrNoteNotFound
}

type mockTokenIssuer struct {
	issueFn func(roomID, memberID string) (string, error)
	parseFn func(token string) (ports.Claims, error)
}

func (m *mockTokenIssuer) Issue(roomID, memberID string) (string, error) {
	if m.issueFn != nil {
		return m.issueFn(roomID, memberID)
	}
	return "mock-token", nil
}

func (m *mockTokenIssuer) Parse(token string) (ports.Claims, error) {
	if m.parseFn != nil {
		return m.parseFn(token)
	}
	return ports.Claims{}, domain.ErrUnauthorized
}

type mockBroadcaster struct {
	mu      sync.Mutex
	events  []ports.Event
	roomIDs []string
}

func (m *mockBroadcaster) Broadcast(roomID string, ev ports.Event) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, ev)
	m.roomIDs = append(m.roomIDs, roomID)
}

func (m *mockBroadcaster) lastEvent() (string, ports.Event, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.events) == 0 {
		return "", ports.Event{}, false
	}
	i := len(m.events) - 1
	return m.roomIDs[i], m.events[i], true
}

func (m *mockBroadcaster) eventCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.events)
}
