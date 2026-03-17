package stream

import "sync"


// StreamPacket encapsulates the raw H.264/H.265 payload and its metadata.
type StreamPacket struct {
	IsKeyFrame bool
	Payload    []byte
	MediaType  uint8  // 0 = Video, 1 = Audio
	CodecID    uint32
}

// MediaType constants to match your C++ definitions
const (
	MediaTypeVideo uint8 = 0
	MediaTypeAudio uint8 = 1
)

// Hub maintains the set of active clients and broadcasts video frames.
type Hub struct {
	clients    map[*Client]bool
	Broadcast  chan StreamPacket
	Register   chan *Client
	Unregister chan *Client
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		// Broadcast:  make(chan StreamPacket, 999999), // Buffered to handle high FPS bursts
		Broadcast:  make(chan StreamPacket, 256), // Buffered to handle high FPS bursts
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
		case packet := <-h.Broadcast:
			for client := range h.clients {
				// Late Joiner Logic: Drop frames until the first IDR Keyframe arrives
				if client.waitingForKeyframe {
					if !packet.IsKeyFrame {
						continue
					}
					client.waitingForKeyframe = false
				}

				select {
				case client.send <- payloadForWebSocket(packet):
				default:
					// Slow consumer detected: drop the client to prevent blocking the Hub
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}