package ws

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"zametka/internal/ports"
)

func startHub(t *testing.T) (*Hub, context.CancelFunc) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	hub := NewHub()
	go hub.Run(ctx)
	t.Cleanup(cancel)
	return hub, cancel
}

func waitHub(t *testing.T) {
	t.Helper()
	time.Sleep(20 * time.Millisecond)
}

func testClient(hub *Hub, roomID string) *Client {
	return newClient(hub, nil, roomID)
}

func recvMessage(t *testing.T, c *Client, timeout time.Duration) []byte {
	t.Helper()
	select {
	case msg := <-c.outbox:
		return msg
	case <-time.After(timeout):
		t.Fatal("timed out waiting for message")
		return nil
	}
}

func TestHub_BroadcastRoomIsolation(t *testing.T) {
	t.Parallel()

	hub, _ := startHub(t)
	clientA := testClient(hub, "room-a")
	clientB := testClient(hub, "room-b")

	hub.Register(clientA)
	hub.Register(clientB)
	waitHub(t)

	hub.Broadcast("room-a", ports.Event{Type: EventNoteCreated, Data: map[string]string{"id": "1"}})
	waitHub(t)

	msg := recvMessage(t, clientA, time.Second)
	var ev ports.Event
	if err := json.Unmarshal(msg, &ev); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if ev.Type != EventNoteCreated {
		t.Errorf("event type = %q, want %q", ev.Type, EventNoteCreated)
	}

	select {
	case <-clientB.outbox:
		t.Fatal("room-b client should not receive room-a broadcast")
	case <-time.After(50 * time.Millisecond):
	}
}

func TestHub_DisconnectCleansMap(t *testing.T) {
	t.Parallel()

	hub, _ := startHub(t)
	client := testClient(hub, "room-1")
	hub.Register(client)
	waitHub(t)

	hub.Broadcast("room-1", ports.Event{Type: EventNoteCreated, Data: "ping"})
	waitHub(t)
	recvMessage(t, client, time.Second)

	hub.Unregister(client)
	waitHub(t)

	hub.Broadcast("room-1", ports.Event{Type: EventNoteUpdated, Data: "after"})
	waitHub(t)

	select {
	case <-client.outbox:
		t.Fatal("unregistered client should not receive broadcasts")
	case <-time.After(50 * time.Millisecond):
	}

	replacement := testClient(hub, "room-1")
	hub.Register(replacement)
	waitHub(t)

	hub.Broadcast("room-1", ports.Event{Type: EventNoteCreated, Data: "fresh"})
	waitHub(t)
	recvMessage(t, replacement, time.Second)
}

func TestHub_SlowConsumerDoesNotBlockHub(t *testing.T) {
	t.Parallel()

	hub, _ := startHub(t)
	slow := testClient(hub, "room-1")

	for i := 0; i < sendBufferSize; i++ {
		slow.outbox <- []byte("prefill")
	}
	hub.Register(slow)
	waitHub(t)

	done := make(chan struct{})
	go func() {
		hub.Broadcast("room-1", ports.Event{Type: EventNoteCreated, Data: "overflow"})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Broadcast blocked on slow consumer")
	}

	waitHub(t)
}

func TestHub_ConcurrentBroadcast(t *testing.T) {
	t.Parallel()

	hub, _ := startHub(t)
	const clients = 4
	const broadcasts = 32

	var wg sync.WaitGroup
	receivers := make([]*Client, clients)
	for i := range receivers {
		c := testClient(hub, "room-concurrent")
		receivers[i] = c
		hub.Register(c)
	}
	waitHub(t)

	for i := 0; i < broadcasts; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			hub.Broadcast("room-concurrent", ports.Event{
				Type: EventNoteCreated,
				Data: n,
			})
		}(i)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("concurrent Broadcast calls blocked")
	}
}

func TestHub_BroadcastNonBlockingWhenHubQueueFull(t *testing.T) {
	t.Parallel()

	hub := NewHub()

	for i := 0; i < cap(hub.broadcast)+1; i++ {
		hub.Broadcast("room", ports.Event{Type: "test", Data: i})
	}

	done := make(chan struct{})
	go func() {
		hub.Broadcast("room", ports.Event{Type: "test", Data: "drop"})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Broadcast blocked when hub queue is full")
	}
}
