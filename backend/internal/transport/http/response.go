package http

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"zametka/internal/domain"
)

type errorBody struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeError(c *fiber.Ctx, err error) error {
	status, code := mapError(err)
	return c.Status(status).JSON(errorBody{
		Error: errorDetail{
			Code:    code,
			Message: err.Error(),
		},
	})
}

func mapError(err error) (int, string) {
	switch {
	case errors.Is(err, domain.ErrValidation):
		return fiber.StatusBadRequest, "VALIDATION"
	case errors.Is(err, domain.ErrUnauthorized):
		return fiber.StatusUnauthorized, "UNAUTHORIZED"
	case errors.Is(err, domain.ErrForbidden):
		return fiber.StatusForbidden, "FORBIDDEN"
	case errors.Is(err, domain.ErrRoomNotFound):
		return fiber.StatusNotFound, "ROOM_NOT_FOUND"
	case errors.Is(err, domain.ErrNoteNotFound):
		return fiber.StatusNotFound, "NOTE_NOT_FOUND"
	case errors.Is(err, domain.ErrRoomFull):
		return fiber.StatusConflict, "ROOM_FULL"
	case errors.Is(err, domain.ErrCodeTaken):
		return fiber.StatusConflict, "CODE_TAKEN"
	default:
		return fiber.StatusInternalServerError, "INTERNAL"
	}
}

func ErrorHandler(c *fiber.Ctx, err error) error {
	if fiberErr, ok := err.(*fiber.Error); ok {
		return c.Status(fiberErr.Code).JSON(errorBody{
			Error: errorDetail{
				Code:    "INTERNAL",
				Message: fiberErr.Message,
			},
		})
	}
	return writeError(c, err)
}
