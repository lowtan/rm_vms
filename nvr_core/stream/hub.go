package stream

import "sync"

// VideoFrame encapsulates the raw H.264/H.265 payload and its metadata.
type VideoFrame struct {
	IsKeyFrame bool
	Payload    []byte
}

// Hub maintains the set of active clients and broadcasts video frames.
type Hub struct {
	clients    map[*Client]bool
	Broadcast  chan VideoFrame
	Register   chan *Client
	Unregister chan *Client
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		// Broadcast:  make(chan VideoFrame, 999999), // Buffered to handle high FPS bursts
		Broadcast:  make(chan VideoFrame, 256), // Buffered to handle high FPS bursts
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.clients[client] = true
		case client := <-h.Unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case frame := <-h.Broadcast:
			for client := range h.clients {
				// Late Joiner Logic: Drop frames until the first IDR Keyframe arrives
				if client.waitingForKeyframe {
					if !frame.IsKeyFrame {
						continue
					}
					client.waitingForKeyframe = false
				}

				select {
				case client.send <- frame.Payload:
				default:
					// Slow consumer detected: drop the client to prevent blocking the Hub
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}