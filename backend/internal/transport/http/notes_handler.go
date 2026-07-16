package http

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"zametka/internal/auth"
	"zametka/internal/domain"
	"zametka/internal/ports"
)

type NotesHandler struct {
	svc ports.NoteService
}

func NewNotesHandler(svc ports.NoteService) *NotesHandler {
	return &NotesHandler{svc: svc}
}

func (h *NotesHandler) List(c *fiber.Ctx) error {
	roomID := auth.RoomID(c)

	filter := domain.NoteFilter{
		RoomID: roomID,
	}

	if cat := c.Query("category"); cat != "" {
		category := domain.Category(cat)
		if !category.Valid() {
			return domain.ErrValidation
		}
		filter.Category = &category
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		limit, err := strconv.ParseInt(limitStr, 10, 64)
		if err != nil || limit <= 0 {
			return domain.ErrValidation
		}
		filter.Limit = limit
	}

	if beforeStr := c.Query("before"); beforeStr != "" {
		before, err := time.Parse(time.RFC3339, beforeStr)
		if err != nil {
			return domain.ErrValidation
		}
		filter.Before = &before
	}

	notes, err := h.svc.List(c.Context(), roomID, filter)
	if err != nil {
		return err
	}

	resp := listNotesResponse{Notes: notes}
	if len(notes) > 0 {
		limit := filter.Limit
		if limit <= 0 {
			limit = 50
		}
		if int64(len(notes)) == limit {
			last := notes[len(notes)-1].CreatedAt
			resp.NextBefore = &last
		}
	}

	return c.JSON(resp)
}

func (h *NotesHandler) Create(c *fiber.Ctx) error {
	var req createNoteRequest
	if err := c.BodyParser(&req); err != nil {
		return domain.ErrValidation
	}

	note, err := h.svc.Create(c.Context(), auth.RoomID(c), auth.MemberID(c), domain.NoteCreate{
		Title:    req.Title,
		Content:  req.Content,
		Category: req.Category,
		Color:    req.Color,
		Pinned:   req.Pinned,
	})
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(note)
}

func (h *NotesHandler) Update(c *fiber.Ctx) error {
	var req updateNoteRequest
	if err := c.BodyParser(&req); err != nil {
		return domain.ErrValidation
	}

	note, err := h.svc.Update(c.Context(), auth.RoomID(c), auth.MemberID(c), c.Params("id"), domain.NoteUpdate{
		Title:    req.Title,
		Content:  req.Content,
		Category: req.Category,
		Color:    req.Color,
		Pinned:   req.Pinned,
	})
	if err != nil {
		return err
	}

	return c.JSON(note)
}

func (h *NotesHandler) Delete(c *fiber.Ctx) error {
	if err := h.svc.Delete(c.Context(), auth.RoomID(c), auth.MemberID(c), c.Params("id")); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}
