package ws

import (
	"context"
	"encoding/json"
	"sync"

	"zametka/internal/ports"
)

const sendBufferSize = 64

type Hub struct {
	mu         sync.RWMutex
	rooms      map[string]map[*Client]struct{}
	register   chan *Client
	unregister chan *Client
	broadcast  chan broadcastMsg
}

type broadcastMsg struct {
	roomID string
	ev     ports.Event
}

func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]struct{}),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan broadcastMsg, 256),
	}
}

func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case client := <-h.register:
			h.mu.Lock()
			if h.rooms[client.roomID] == nil {
				h.rooms[client.roomID] = make(map[*Client]struct{})
			}
			h.rooms[client.roomID][client] = struct{}{}
			h.mu.Unlock()
		case client := <-h.unregister:
			h.removeClient(client)
		case msg := <-h.broadcast:
			h.deliver(msg.roomID, msg.ev)
		}
	}
}

func (h *Hub) Broadcast(roomID string, ev ports.Event) {
	select {
	case h.broadcast <- broadcastMsg{roomID: roomID, ev: ev}:
	default:
		// Drop if hub is overloaded; do not block callers.
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

func (h *Hub) removeClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	roomClients, ok := h.rooms[client.roomID]
	if !ok {
		return
	}
	delete(roomClients, client)
	if len(roomClients) == 0 {
		delete(h.rooms, client.roomID)
	}
}

func (h *Hub) deliver(roomID string, ev ports.Event) {
	payload, err := json.Marshal(ev)
	if err != nil {
		return
	}

	h.mu.RLock()
	roomClients := h.rooms[roomID]
	clients := make([]*Client, 0, len(roomClients))
	for c := range roomClients {
		clients = append(clients, c)
	}
	h.mu.RUnlock()

	for _, client := range clients {
		client.send(payload)
	}
}

var _ ports.Broadcaster = (*Hub)(nil)
