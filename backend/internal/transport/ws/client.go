package ws

import (
	"sync"

	"github.com/gofiber/contrib/websocket"
)

type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	roomID string
	outbox chan []byte
	once   sync.Once
}

func newClient(hub *Hub, conn *websocket.Conn, roomID string) *Client {
	return &Client{
		hub:    hub,
		conn:   conn,
		roomID: roomID,
		outbox: make(chan []byte, sendBufferSize),
	}
}

func (c *Client) writePump() {
	defer c.close()

	for msg := range c.outbox {
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}

func (c *Client) send(msg []byte) {
	select {
	case c.outbox <- msg:
	default:
		c.close()
	}
}

func (c *Client) close() {
	c.once.Do(func() {
		close(c.outbox)
		if c.conn != nil {
			_ = c.conn.Close()
		}
		c.hub.Unregister(c)
	})
}
