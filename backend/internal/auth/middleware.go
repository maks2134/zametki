package auth

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"zametka/internal/domain"
	"zametka/internal/ports"
)

const (
	localRoomID   = "roomID"
	localMemberID = "memberID"
)

func RequireAuth(issuer ports.TokenIssuer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := extractToken(c)
		if token == "" {
			return domain.ErrUnauthorized
		}

		claims, err := issuer.Parse(token)
		if err != nil {
			return err
		}

		c.Locals(localRoomID, claims.RoomID)
		c.Locals(localMemberID, claims.MemberID)
		return c.Next()
	}
}

func extractToken(c *fiber.Ctx) string {
	if c.Path() == "/ws" {
		return c.Query("token")
	}

	auth := c.Get("Authorization")
	if auth == "" {
		return ""
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func RoomID(c *fiber.Ctx) string {
	v, _ := c.Locals(localRoomID).(string)
	return v
}

func MemberID(c *fiber.Ctx) string {
	v, _ := c.Locals(localMemberID).(string)
	return v
}
