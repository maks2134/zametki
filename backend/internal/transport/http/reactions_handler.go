package http

import (
	"github.com/gofiber/fiber/v2"

	"zametka/internal/auth"
	"zametka/internal/domain"
	"zametka/internal/ports"
)

type ReactionsHandler struct {
	svc ports.NoteService
}

func NewReactionsHandler(svc ports.NoteService) *ReactionsHandler {
	return &ReactionsHandler{svc: svc}
}

func (h *ReactionsHandler) Add(c *fiber.Ctx) error {
	var req addReactionRequest
	if err := c.BodyParser(&req); err != nil {
		return domain.ErrValidation
	}

	note, err := h.svc.AddReaction(c.Context(), auth.RoomID(c), auth.MemberID(c), c.Params("id"), req.Emoji)
	if err != nil {
		return err
	}

	return c.JSON(note)
}

func (h *ReactionsHandler) Remove(c *fiber.Ctx) error {
	note, err := h.svc.RemoveReaction(c.Context(), auth.RoomID(c), auth.MemberID(c), c.Params("id"))
	if err != nil {
		return err
	}

	return c.JSON(note)
}
