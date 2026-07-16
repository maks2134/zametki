package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"zametka/internal/domain"
	"zametka/internal/ports"
)

const issuerName = "zametka"

type TokenIssuer struct {
	secret []byte
	ttl    time.Duration
}

func NewTokenIssuer(secret string, ttl time.Duration) *TokenIssuer {
	return &TokenIssuer{
		secret: []byte(secret),
		ttl:    ttl,
	}
}

type claims struct {
	RoomID string `json:"rid"`
	jwt.RegisteredClaims
}

func (t *TokenIssuer) Issue(roomID, memberID string) (string, error) {
	now := time.Now().UTC()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		RoomID: roomID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   memberID,
			Issuer:    issuerName,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(t.ttl)),
		},
	})

	signed, err := token.SignedString(t.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

func (t *TokenIssuer) Parse(tokenStr string) (ports.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &claims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return t.secret, nil
	})
	if err != nil {
		return ports.Claims{}, domain.ErrUnauthorized
	}

	c, ok := token.Claims.(*claims)
	if !ok || !token.Valid {
		return ports.Claims{}, domain.ErrUnauthorized
	}
	if c.Issuer != issuerName || c.Subject == "" || c.RoomID == "" {
		return ports.Claims{}, domain.ErrUnauthorized
	}

	return ports.Claims{
		RoomID:   c.RoomID,
		MemberID: c.Subject,
	}, nil
}

var _ ports.TokenIssuer = (*TokenIssuer)(nil)
