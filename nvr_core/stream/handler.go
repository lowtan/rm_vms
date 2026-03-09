package stream

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait = 10 * time.Second
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024 * 64, // 64KB for large video keyframes
	CheckOrigin: func(r *http.Request) bool {
		return true // Lock this down in production
	},
}

type Client struct {
	hub                *Hub
	conn               *websocket.Conn
	send               chan []byte
	waitingForKeyframe bool
}

// ServeWs handles websocket requests from the Vue frontend.
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket Upgrade Error: %v", err)
		return
	}

	client := &Client{
		hub:                hub,
		conn:               conn,
		send:               make(chan []byte, 256),
		waitingForKeyframe: true, // Crucial for clean jmuxer initialization
	}
	client.hub.Register <- client

	// Start the write pump for binary frames
	go client.writePump()
	
	// Start the read pump to handle client disconnects gracefully
	go client.readPump()
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()
	for {
		select {
		case payload, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.BinaryMessage)
			if err != nil {
				return
			}
			w.Write(payload)

			if err := w.Close(); err != nil {
				return
			}
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.Unregister <- c
		c.conn.Close()
	}()
	for {
		// SPSC stream: We only read to detect close events from the browser
		if _, _, err := c.conn.ReadMessage(); err != nil {
			break
		}
	}
}