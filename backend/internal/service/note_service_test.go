package service

import (
	"context"
	"errors"
	"testing"

	"zametka/internal/domain"
)

func sampleNote(authorID string) *domain.Note {
	return &domain.Note{
		ID:       "note-1",
		RoomID:   "room-1",
		AuthorID: authorID,
		Content:  "hello",
		Category: domain.CategoryGift,
	}
}

func TestNoteService_CreateValidation(t *testing.T) {
	t.Parallel()

	svc := NewNoteService(&mockNoteRepo{}, &mockBroadcaster{})

	tests := []struct {
		name string
		in   domain.NoteCreate
	}{
		{
			name: "empty content",
			in:   domain.NoteCreate{Content: "  ", Category: domain.CategoryGift},
		},
		{
			name: "invalid category",
			in:   domain.NoteCreate{Content: "hello", Category: domain.Category("bad")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := svc.Create(context.Background(), "room-1", "author", tt.in)
			if !errors.Is(err, domain.ErrValidation) {
				t.Errorf("Create() error = %v, want %v", err, domain.ErrValidation)
			}
		})
	}
}

func TestNoteService_UpdateForbidden(t *testing.T) {
	t.Parallel()

	repo := &mockNoteRepo{
		getByIDFn: func(_ context.Context, _, _ string) (*domain.Note, error) {
			return sampleNote("author-1"), nil
		},
	}
	svc := NewNoteService(repo, &mockBroadcaster{})

	content := "updated"
	_, err := svc.Update(context.Background(), "room-1", "other-member", "note-1", domain.NoteUpdate{
		Content: &content,
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("Update() error = %v, want %v", err, domain.ErrForbidden)
	}
}

func TestNoteService_UpdateByAuthor(t *testing.T) {
	t.Parallel()

	updated := sampleNote("author-1")
	updated.Content = "updated"
	repo := &mockNoteRepo{
		getByIDFn: func(_ context.Context, _, _ string) (*domain.Note, error) {
			return sampleNote("author-1"), nil
		},
		updateFn: func(_ context.Context, _, _ string, _ domain.NoteUpdate) (*domain.Note, error) {
			return updated, nil
		},
	}
	bc := &mockBroadcaster{}
	svc := NewNoteService(repo, bc)

	content := "updated"
	note, err := svc.Update(context.Background(), "room-1", "author-1", "note-1", domain.NoteUpdate{
		Content: &content,
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if note.Content != "updated" {
		t.Errorf("note.Content = %q, want updated", note.Content)
	}

	roomID, ev, ok := bc.lastEvent()
	if !ok {
		t.Fatal("expected broadcast event")
	}
	if roomID != "room-1" {
		t.Errorf("broadcast roomID = %q, want room-1", roomID)
	}
	if ev.Type != "note.updated" {
		t.Errorf("event type = %q, want note.updated", ev.Type)
	}
}

func TestNoteService_DeleteForbidden(t *testing.T) {
	t.Parallel()

	repo := &mockNoteRepo{
		getByIDFn: func(_ context.Context, _, _ string) (*domain.Note, error) {
			return sampleNote("author-1"), nil
		},
	}
	svc := NewNoteService(repo, &mockBroadcaster{})

	err := svc.Delete(context.Background(), "room-1", "other-member", "note-1")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("Delete() error = %v, want %v", err, domain.ErrForbidden)
	}
}

func TestNoteService_DeleteByAuthor(t *testing.T) {
	t.Parallel()

	repo := &mockNoteRepo{
		getByIDFn: func(_ context.Context, _, _ string) (*domain.Note, error) {
			return sampleNote("author-1"), nil
		},
		deleteFn: func(_ context.Context, _, _ string) error {
			return nil
		},
	}
	bc := &mockBroadcaster{}
	svc := NewNoteService(repo, bc)

	err := svc.Delete(context.Background(), "room-1", "author-1", "note-1")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	roomID, ev, ok := bc.lastEvent()
	if !ok {
		t.Fatal("expected broadcast event")
	}
	if roomID != "room-1" {
		t.Errorf("broadcast roomID = %q, want room-1", roomID)
	}
	if ev.Type != "note.deleted" {
		t.Errorf("event type = %q, want note.deleted", ev.Type)
	}
	data, ok := ev.Data.(map[string]string)
	if !ok {
		t.Fatalf("event data type = %T, want map[string]string", ev.Data)
	}
	if data["id"] != "note-1" {
		t.Errorf("deleted id = %q, want note-1", data["id"])
	}
}

func TestNoteService_AddReactionBroadcasts(t *testing.T) {
	t.Parallel()

	note := sampleNote("author-1")
	repo := &mockNoteRepo{
		upsertReactionFn: func(_ context.Context, _, _, _, _ string) (*domain.Note, error) {
			return note, nil
		},
	}
	bc := &mockBroadcaster{}
	svc := NewNoteService(repo, bc)

	_, err := svc.AddReaction(context.Background(), "room-1", "member-2", "note-1", "❤️")
	if err != nil {
		t.Fatalf("AddReaction() error = %v", err)
	}

	_, ev, ok := bc.lastEvent()
	if !ok {
		t.Fatal("expected broadcast event")
	}
	if ev.Type != "reaction.updated" {
		t.Errorf("event type = %q, want reaction.updated", ev.Type)
	}
}

func TestNoteService_RemoveReactionBroadcasts(t *testing.T) {
	t.Parallel()

	note := sampleNote("author-1")
	repo := &mockNoteRepo{
		removeReactionFn: func(_ context.Context, _, _, _ string) (*domain.Note, error) {
			return note, nil
		},
	}
	bc := &mockBroadcaster{}
	svc := NewNoteService(repo, bc)

	_, err := svc.RemoveReaction(context.Background(), "room-1", "member-2", "note-1")
	if err != nil {
		t.Fatalf("RemoveReaction() error = %v", err)
	}

	_, ev, ok := bc.lastEvent()
	if !ok {
		t.Fatal("expected broadcast event")
	}
	if ev.Type != "reaction.updated" {
		t.Errorf("event type = %q, want reaction.updated", ev.Type)
	}
}

func TestNoteService_CreateSuccess(t *testing.T) {
	t.Parallel()

	var created *domain.Note
	repo := &mockNoteRepo{
		createFn: func(_ context.Context, n *domain.Note) error {
			created = n
			return nil
		},
	}
	bc := &mockBroadcaster{}
	svc := NewNoteService(repo, bc)

	note, err := svc.Create(context.Background(), "room-1", "author-1", domain.NoteCreate{
		Content:  "hello",
		Category: domain.CategoryGift,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if note != created {
		t.Error("Create() returned different note than repo received")
	}
	if note.Reactions == nil || len(note.Reactions) != 0 {
		t.Errorf("Reactions = %v, want empty slice", note.Reactions)
	}

	_, ev, ok := bc.lastEvent()
	if !ok {
		t.Fatal("expected broadcast event")
	}
	if ev.Type != "note.created" {
		t.Errorf("event type = %q, want note.created", ev.Type)
	}
}
