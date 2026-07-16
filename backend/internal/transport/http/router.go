package http

import (
	"github.com/gofiber/fiber/v2"

	"zametka/internal/auth"
	"zametka/internal/ports"
)

func RegisterRoutes(app *fiber.App, roomSvc ports.RoomService, noteSvc ports.NoteService, issuer ports.TokenIssuer) {
	rooms := NewRoomsHandler(roomSvc)
	notes := NewNotesHandler(noteSvc)
	reactions := NewReactionsHandler(noteSvc)

	api := app.Group("/api")

	api.Post("/rooms", rooms.Create)
	api.Post("/rooms/join", rooms.Join)

	protected := api.Group("", auth.RequireAuth(issuer))
	protected.Get("/rooms/me", rooms.Me)

	protected.Get("/notes", notes.List)
	protected.Post("/notes", notes.Create)
	protected.Patch("/notes/:id", notes.Update)
	protected.Delete("/notes/:id", notes.Delete)
	protected.Post("/notes/:id/reactions", reactions.Add)
	protected.Delete("/notes/:id/reactions", reactions.Remove)
}
