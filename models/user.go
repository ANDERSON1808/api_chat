package models

import (
	"time"

	"github.com/gorilla/websocket"
)

const (
	// WriteWait Time allowed to write a message to the peer.
	WriteWait = 10 * time.Second
	// PongWait Time allowed to read the next pong message from the peer.
	PongWait = 60 * time.Second
	// PingPeriod Send pings to peer with this period. Must be less than pongWait.
	PingPeriod = (PongWait * 9) / 10
	// MaxMessageSize Maximum message size allowed from peer.
	MaxMessageSize = 512
)

// Client represents a user in a ChatRoom
type Client struct {
	Username     string    `json:"username"`
	Color        string    `json:"color"`
	LastActivity time.Time `json:"last_activity"`
	// The websocket Connection.
	Conn *websocket.Conn `json:"-"`
	// Buffered channel of outbound messages.
	Send chan []byte `json:"-"`
	// ChatRoom that client is registered with
	Room *ChatRoom `json:"-"`
}
