package http

import (
	"time"

	"zametka/internal/domain"
)

type createRoomRequest struct {
	Title string `json:"title"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type joinRoomRequest struct {
	Code  string `json:"code"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type createRoomResponse struct {
	Room  domain.Room `json:"room"`
	Token string      `json:"token"`
	Code  string      `json:"code"`
}

type joinRoomResponse struct {
	Room  domain.Room `json:"room"`
	Token string      `json:"token"`
}

type roomResponse struct {
	Room domain.Room `json:"room"`
}

type createNoteRequest struct {
	Title    string          `json:"title"`
	Content  string          `json:"content"`
	Category domain.Category `json:"category"`
	Color    string          `json:"color"`
	Pinned   bool            `json:"pinned"`
}

type updateNoteRequest struct {
	Title    *string          `json:"title"`
	Content  *string          `json:"content"`
	Category *domain.Category `json:"category"`
	Color    *string          `json:"color"`
	Pinned   *bool            `json:"pinned"`
}

type listNotesResponse struct {
	Notes      []domain.Note `json:"notes"`
	NextBefore *time.Time    `json:"nextBefore,omitempty"`
}

type addReactionRequest struct {
	Emoji string `json:"emoji"`
}
