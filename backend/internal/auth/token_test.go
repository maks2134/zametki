package auth

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"zametka/internal/domain"
)

const testSecret = "change-me-min-32-bytes-long-secret!!"

func TestTokenIssuer_IssueParseRoundTrip(t *testing.T) {
	t.Parallel()

	issuer := NewTokenIssuer(testSecret, time.Hour)
	roomID := "room-abc"
	memberID := "member-xyz"

	token, err := issuer.Issue(roomID, memberID)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	if token == "" {
		t.Fatal("Issue() returned empty token")
	}

	claims, err := issuer.Parse(token)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if claims.RoomID != roomID {
		t.Errorf("RoomID = %q, want %q", claims.RoomID, roomID)
	}
	if claims.MemberID != memberID {
		t.Errorf("MemberID = %q, want %q", claims.MemberID, memberID)
	}
}

func TestTokenIssuer_ParseErrors(t *testing.T) {
	t.Parallel()

	issuer := NewTokenIssuer(testSecret, time.Hour)

	tests := []struct {
		name  string
		token func(t *testing.T) string
	}{
		{
			name: "expired",
			token: func(t *testing.T) string {
				t.Helper()
				short := NewTokenIssuer(testSecret, time.Millisecond)
				tok, err := short.Issue("room", "member")
				if err != nil {
					t.Fatalf("Issue() error = %v", err)
				}
				time.Sleep(5 * time.Millisecond)
				return tok
			},
		},
		{
			name: "tampered signature",
			token: func(t *testing.T) string {
				t.Helper()
				tok, err := issuer.Issue("room", "member")
				if err != nil {
					t.Fatalf("Issue() error = %v", err)
				}
				parts := strings.Split(tok, ".")
				if len(parts) != 3 {
					t.Fatalf("unexpected token format: %q", tok)
				}
				last := parts[2]
				if last[len(last)-1] == 'a' {
					parts[2] = last[:len(last)-1] + "b"
				} else {
					parts[2] = last[:len(last)-1] + "a"
				}
				return strings.Join(parts, ".")
			},
		},
		{
			name: "wrong signing algorithm",
			token: func(t *testing.T) string {
				t.Helper()
				now := time.Now().UTC()
				token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims{
					RoomID: "room",
					RegisteredClaims: jwt.RegisteredClaims{
						Subject:   "member",
						Issuer:    issuerName,
						IssuedAt:  jwt.NewNumericDate(now),
						ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
					},
				})
				signed, err := token.SignedString([]byte(testSecret))
				if err != nil {
					t.Fatalf("SignedString() error = %v", err)
				}
				return signed
			},
		},
		{
			name: "wrong secret",
			token: func(t *testing.T) string {
				t.Helper()
				other := NewTokenIssuer("another-secret-at-least-32-bytes!!", time.Hour)
				tok, err := other.Issue("room", "member")
				if err != nil {
					t.Fatalf("Issue() error = %v", err)
				}
				return tok
			},
		},
		{
			name: "garbage",
			token: func(t *testing.T) string {
				return "not.a.jwt"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := issuer.Parse(tt.token(t))
			if !errors.Is(err, domain.ErrUnauthorized) {
				t.Errorf("Parse() error = %v, want %v", err, domain.ErrUnauthorized)
			}
		})
	}
}
