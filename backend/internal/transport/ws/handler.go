package ws

import (
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"

	"zametka/internal/ports"
)

func RegisterWS(app *fiber.App, hub *Hub, issuer ports.TokenIssuer) {
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws", websocket.New(func(conn *websocket.Conn) {
		token := conn.Query("token")
		claims, err := issuer.Parse(token)
		if err != nil {
			_ = conn.Close()
			return
		}

		client := newClient(hub, conn, claims.RoomID)
		hub.Register(client)
		go client.writePump()

		defer client.close()

		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}))
}
