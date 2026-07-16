package service

import (
	"context"
	"errors"
	"testing"

	"zametka/internal/domain"
)

func TestRoomService_Create(t *testing.T) {
	t.Parallel()

	bc := &mockBroadcaster{}
	var created *domain.Room
	repo := &mockRoomRepo{
		createFn: func(_ context.Context, room *domain.Room) error {
			created = room
			return nil
		},
	}
	issuer := &mockTokenIssuer{
		issueFn: func(roomID, memberID string) (string, error) {
			if roomID == "" || memberID == "" {
				t.Errorf("Issue called with empty ids: room=%q member=%q", roomID, memberID)
			}
			return "jwt-token", nil
		},
	}

	svc := NewRoomService(repo, issuer, bc)
	room, token, err := svc.Create(context.Background(), "Our Notes", "Alex", "#f28b82")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if room != created {
		t.Fatal("Create() returned different room pointer than repo received")
	}
	if room.ID == "" {
		t.Error("room.ID is empty")
	}
	if len(room.Code) != joinCodeLength {
		t.Errorf("room.Code length = %d, want %d", len(room.Code), joinCodeLength)
	}
	if room.Title != "Our Notes" {
		t.Errorf("room.Title = %q, want %q", room.Title, "Our Notes")
	}
	if len(room.Members) != 1 {
		t.Fatalf("len(room.Members) = %d, want 1", len(room.Members))
	}
	if room.Members[0].Name != "Alex" {
		t.Errorf("member name = %q, want Alex", room.Members[0].Name)
	}
	if token != "jwt-token" {
		t.Errorf("token = %q, want jwt-token", token)
	}
	if bc.eventCount() != 0 {
		t.Error("Create should not broadcast events")
	}
}

func TestRoomService_JoinRoomFull(t *testing.T) {
	t.Parallel()

	repo := &mockRoomRepo{
		addMemberFn: func(_ context.Context, _ string, _ domain.Member) (*domain.Room, error) {
			return nil, domain.ErrRoomFull
		},
	}
	svc := NewRoomService(repo, &mockTokenIssuer{}, &mockBroadcaster{})

	_, _, err := svc.Join(context.Background(), "ABC123", "Sam", "#a7ffeb")
	if !errors.Is(err, domain.ErrRoomFull) {
		t.Errorf("Join() error = %v, want %v", err, domain.ErrRoomFull)
	}
}

func TestRoomService_JoinBroadcastsMemberJoined(t *testing.T) {
	t.Parallel()

	existingRoom := &domain.Room{
		ID:   "room-1",
		Code: "K7M4QP",
	}
	var addedMember domain.Member
	repo := &mockRoomRepo{
		addMemberFn: func(_ context.Context, code string, m domain.Member) (*domain.Room, error) {
			if code != "K7M4QP" {
				t.Errorf("AddMember code = %q, want K7M4QP", code)
			}
			addedMember = m
			return existingRoom, nil
		},
	}
	issuer := &mockTokenIssuer{
		issueFn: func(roomID, memberID string) (string, error) {
			if roomID != existingRoom.ID {
				t.Errorf("Issue roomID = %q, want %q", roomID, existingRoom.ID)
			}
			if memberID != addedMember.ID {
				t.Errorf("Issue memberID = %q, want %q", memberID, addedMember.ID)
			}
			return "join-token", nil
		},
	}
	bc := &mockBroadcaster{}
	svc := NewRoomService(repo, issuer, bc)

	room, token, err := svc.Join(context.Background(), "k7m4qp", "Sam", "#a7ffeb")
	if err != nil {
		t.Fatalf("Join() error = %v", err)
	}
	if room != existingRoom {
		t.Error("Join() returned unexpected room")
	}
	if token != "join-token" {
		t.Errorf("token = %q, want join-token", token)
	}

	roomID, ev, ok := bc.lastEvent()
	if !ok {
		t.Fatal("expected broadcast event")
	}
	if roomID != existingRoom.ID {
		t.Errorf("broadcast roomID = %q, want %q", roomID, existingRoom.ID)
	}
	if ev.Type != "member.joined" {
		t.Errorf("event type = %q, want member.joined", ev.Type)
	}
	member, ok := ev.Data.(domain.Member)
	if !ok {
		t.Fatalf("event data type = %T, want domain.Member", ev.Data)
	}
	if member.ID != addedMember.ID || member.Name != "Sam" {
		t.Errorf("broadcast member = %+v, want id=%s name=Sam", member, addedMember.ID)
	}
}

func TestRoomService_CreateValidation(t *testing.T) {
	t.Parallel()

	svc := NewRoomService(&mockRoomRepo{}, &mockTokenIssuer{}, &mockBroadcaster{})

	tests := []struct {
		name       string
		title      string
		memberName string
		color      string
	}{
		{name: "empty title", title: "", memberName: "Alex", color: "#fff"},
		{name: "empty name", title: "Notes", memberName: "  ", color: "#fff"},
		{name: "empty color", title: "Notes", memberName: "Alex", color: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, _, err := svc.Create(context.Background(), tt.title, tt.memberName, tt.color)
			if !errors.Is(err, domain.ErrValidation) {
				t.Errorf("Create() error = %v, want %v", err, domain.ErrValidation)
			}
		})
	}
}
