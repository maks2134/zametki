package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"

	"zametka/internal/domain"
	"zametka/internal/ports"
)

const joinCodeAlphabet = "ABCDEFGHJKMNPQRSTVWXYZ23456789"
const joinCodeLength = 6
const maxCodeRetries = 5

type RoomService struct {
	repo        ports.RoomRepository
	issuer      ports.TokenIssuer
	broadcaster ports.Broadcaster
}

func NewRoomService(repo ports.RoomRepository, issuer ports.TokenIssuer, broadcaster ports.Broadcaster) *RoomService {
	return &RoomService{
		repo:        repo,
		issuer:      issuer,
		broadcaster: broadcaster,
	}
}

func (s *RoomService) Create(ctx context.Context, title, name, color string) (*domain.Room, string, error) {
	if err := validateMemberInput(name, color); err != nil {
		return nil, "", err
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, "", fmt.Errorf("%w: title is required", domain.ErrValidation)
	}

	now := time.Now().UTC()
	memberID := uuid.NewString()
	room := &domain.Room{
		ID:        uuid.NewString(),
		Title:     title,
		CreatedAt: now,
		Members: []domain.Member{
			{
				ID:       memberID,
				Name:     strings.TrimSpace(name),
				Color:    color,
				JoinedAt: now,
			},
		},
	}

	var lastErr error
	for range maxCodeRetries {
		code, err := generateJoinCode()
		if err != nil {
			return nil, "", err
		}
		room.Code = code

		if err := s.repo.Create(ctx, room); err != nil {
			if errors.Is(err, domain.ErrCodeTaken) {
				lastErr = err
				continue
			}
			return nil, "", err
		}

		token, err := s.issuer.Issue(room.ID, memberID)
		if err != nil {
			return nil, "", err
		}
		return room, token, nil
	}

	if lastErr != nil {
		return nil, "", fmt.Errorf("create room: %w", lastErr)
	}
	return nil, "", fmt.Errorf("create room: failed after %d retries", maxCodeRetries)
}

func (s *RoomService) Join(ctx context.Context, code, name, color string) (*domain.Room, string, error) {
	if err := validateMemberInput(name, color); err != nil {
		return nil, "", err
	}
	code = strings.TrimSpace(strings.ToUpper(code))
	if code == "" {
		return nil, "", fmt.Errorf("%w: code is required", domain.ErrValidation)
	}

	member := domain.Member{
		ID:       uuid.NewString(),
		Name:     strings.TrimSpace(name),
		Color:    color,
		JoinedAt: time.Now().UTC(),
	}

	room, err := s.repo.AddMember(ctx, code, member)
	if err != nil {
		return nil, "", err
	}

	token, err := s.issuer.Issue(room.ID, member.ID)
	if err != nil {
		return nil, "", err
	}

	s.broadcaster.Broadcast(room.ID, ports.Event{
		Type: "member.joined",
		Data: member,
	})

	return room, token, nil
}

func (s *RoomService) Get(ctx context.Context, roomID string) (*domain.Room, error) {
	return s.repo.GetByID(ctx, roomID)
}

func validateMemberInput(name, color string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("%w: name is required", domain.ErrValidation)
	}
	if strings.TrimSpace(color) == "" {
		return fmt.Errorf("%w: color is required", domain.ErrValidation)
	}
	return nil
}

func generateJoinCode() (string, error) {
	max := big.NewInt(int64(len(joinCodeAlphabet)))
	b := make([]byte, joinCodeLength)
	for i := range b {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", fmt.Errorf("generate join code: %w", err)
		}
		b[i] = joinCodeAlphabet[n.Int64()]
	}
	return string(b), nil
}

var _ ports.RoomService = (*RoomService)(nil)
