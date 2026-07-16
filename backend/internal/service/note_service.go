package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"zametka/internal/domain"
	"zametka/internal/ports"
)

type NoteService struct {
	repo        ports.NoteRepository
	broadcaster ports.Broadcaster
}

func NewNoteService(repo ports.NoteRepository, broadcaster ports.Broadcaster) *NoteService {
	return &NoteService{
		repo:        repo,
		broadcaster: broadcaster,
	}
}

func (s *NoteService) List(ctx context.Context, roomID string, f domain.NoteFilter) ([]domain.Note, error) {
	f.RoomID = roomID
	return s.repo.List(ctx, f)
}

func (s *NoteService) Create(ctx context.Context, roomID, authorID string, in domain.NoteCreate) (*domain.Note, error) {
	if err := validateNoteCreate(in); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	note := &domain.Note{
		ID:        uuid.NewString(),
		RoomID:    roomID,
		AuthorID:  authorID,
		Title:     strings.TrimSpace(in.Title),
		Content:   strings.TrimSpace(in.Content),
		Category:  in.Category,
		Color:     in.Color,
		Pinned:    in.Pinned,
		Reactions: []domain.Reaction{},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.Create(ctx, note); err != nil {
		return nil, err
	}

	s.broadcaster.Broadcast(roomID, ports.Event{
		Type: "note.created",
		Data: note,
	})

	return note, nil
}

func (s *NoteService) Update(ctx context.Context, roomID, memberID, noteID string, upd domain.NoteUpdate) (*domain.Note, error) {
	if err := validateNoteUpdate(upd); err != nil {
		return nil, err
	}

	note, err := s.repo.GetByID(ctx, roomID, noteID)
	if err != nil {
		return nil, err
	}
	if note.AuthorID != memberID {
		return nil, domain.ErrForbidden
	}

	updated, err := s.repo.Update(ctx, roomID, noteID, upd)
	if err != nil {
		return nil, err
	}

	s.broadcaster.Broadcast(roomID, ports.Event{
		Type: "note.updated",
		Data: updated,
	})

	return updated, nil
}

func (s *NoteService) Delete(ctx context.Context, roomID, memberID, noteID string) error {
	note, err := s.repo.GetByID(ctx, roomID, noteID)
	if err != nil {
		return err
	}
	if note.AuthorID != memberID {
		return domain.ErrForbidden
	}

	if err := s.repo.Delete(ctx, roomID, noteID); err != nil {
		return err
	}

	s.broadcaster.Broadcast(roomID, ports.Event{
		Type: "note.deleted",
		Data: map[string]string{"id": noteID},
	})

	return nil
}

func (s *NoteService) AddReaction(ctx context.Context, roomID, memberID, noteID, emoji string) (*domain.Note, error) {
	emoji = strings.TrimSpace(emoji)
	if emoji == "" {
		return nil, fmt.Errorf("%w: emoji is required", domain.ErrValidation)
	}

	note, err := s.repo.UpsertReaction(ctx, roomID, noteID, memberID, emoji)
	if err != nil {
		return nil, err
	}

	s.broadcaster.Broadcast(roomID, ports.Event{
		Type: "reaction.updated",
		Data: note,
	})

	return note, nil
}

func (s *NoteService) RemoveReaction(ctx context.Context, roomID, memberID, noteID string) (*domain.Note, error) {
	note, err := s.repo.RemoveReaction(ctx, roomID, noteID, memberID)
	if err != nil {
		return nil, err
	}

	s.broadcaster.Broadcast(roomID, ports.Event{
		Type: "reaction.updated",
		Data: note,
	})

	return note, nil
}

func validateNoteCreate(in domain.NoteCreate) error {
	if strings.TrimSpace(in.Content) == "" {
		return fmt.Errorf("%w: content is required", domain.ErrValidation)
	}
	if !in.Category.Valid() {
		return fmt.Errorf("%w: invalid category", domain.ErrValidation)
	}
	return nil
}

func validateNoteUpdate(upd domain.NoteUpdate) error {
	if upd.Category != nil && !upd.Category.Valid() {
		return fmt.Errorf("%w: invalid category", domain.ErrValidation)
	}
	if upd.Content != nil && strings.TrimSpace(*upd.Content) == "" {
		return fmt.Errorf("%w: content cannot be empty", domain.ErrValidation)
	}
	if upd.Title == nil && upd.Content == nil && upd.Category == nil && upd.Color == nil && upd.Pinned == nil {
		return fmt.Errorf("%w: at least one field is required", domain.ErrValidation)
	}
	return nil
}

var _ ports.NoteService = (*NoteService)(nil)
