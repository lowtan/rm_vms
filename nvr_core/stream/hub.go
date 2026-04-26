package stream

import "sync"

// Subscriber can be a WebSocket subscriber OR an HTTP response writer
type Subscriber struct {
	Send               chan StreamPacket
	WaitingForKeyframe bool
}

// StreamPacket encapsulates the raw video payload and its metadata.
type StreamPacket struct {
	MediaType  uint8  // 0 = Video, 1 = Audio
	CodecID    uint32
	IsKeyFrame bool
	PTS        int64
	DTS        int64
	Payload    []byte
}

// MediaType constants to match your C++ definitions
const (
	MediaTypeVideo uint8 = 0
	MediaTypeAudio uint8 = 1
)

// Hub maintains the set of active clients and broadcasts video frames.
type Hub struct {
	subscribers    map[*Subscriber]bool
	Broadcast  chan StreamPacket
	Register   chan *Subscriber
	Unregister chan *Subscriber
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		Broadcast:  make(chan StreamPacket, 2048), // Buffered to handle high FPS bursts
		Register:   make(chan *Subscriber),
		Unregister: make(chan *Subscriber),
		subscribers:    make(map[*Subscriber]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case subscriber := <-h.Register:
			h.subscribers[subscriber] = true
		case subscriber := <-h.Unregister:
			if _, ok := h.subscribers[subscriber]; ok {
				delete(h.subscribers, subscriber)
				close(subscriber.Send)
			}
		case packet := <-h.Broadcast:
			for subscriber := range h.subscribers {
				// Late Joiner Logic: Drop frames until the first IDR Keyframe arrives
				if subscriber.WaitingForKeyframe {
					if !packet.IsKeyFrame {
						continue
					}
					subscriber.WaitingForKeyframe = false
				}

				select {
				case subscriber.Send <- packet:
				default:
					// Slow consumer detected: drop the subscriber to prevent blocking the Hub
					close(subscriber.Send)
					delete(h.subscribers, subscriber)
				}
			}
		}
	}
}