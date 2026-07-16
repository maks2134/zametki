package http

import (
	"github.com/gofiber/fiber/v2"

	"zametka/internal/auth"
	"zametka/internal/domain"
	"zametka/internal/ports"
)

type RoomsHandler struct {
	svc ports.RoomService
}

func NewRoomsHandler(svc ports.RoomService) *RoomsHandler {
	return &RoomsHandler{svc: svc}
}

func (h *RoomsHandler) Create(c *fiber.Ctx) error {
	var req createRoomRequest
	if err := c.BodyParser(&req); err != nil {
		return domain.ErrValidation
	}

	room, token, err := h.svc.Create(c.Context(), req.Title, req.Name, req.Color)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(createRoomResponse{
		Room:  *room,
		Token: token,
		Code:  room.Code,
	})
}

func (h *RoomsHandler) Join(c *fiber.Ctx) error {
	var req joinRoomRequest
	if err := c.BodyParser(&req); err != nil {
		return domain.ErrValidation
	}

	room, token, err := h.svc.Join(c.Context(), req.Code, req.Name, req.Color)
	if err != nil {
		return err
	}

	return c.JSON(joinRoomResponse{
		Room:  *room,
		Token: token,
	})
}

func (h *RoomsHandler) Me(c *fiber.Ctx) error {
	roomID := auth.RoomID(c)
	room, err := h.svc.Get(c.Context(), roomID)
	if err != nil {
		return err
	}
	return c.JSON(roomResponse{Room: *room})
}
