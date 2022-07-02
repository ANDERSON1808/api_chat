package models

import (
	"time"
)

const (
	// Subscribe is used to broadcast a message indicating user has joined ChatRoom
	Subscribe = "join"
	// Broadcast is used to broadcast messages to all subscribed users
	Broadcast = "send"
	// Unsubscribe is used to broadcast a message indicating user has left ChatRoom
	Unsubscribe = "leave"
)

// ChatEvent represents a message event in an associated ChatRoom
type ChatEvent struct {
	EventType string    `json:"event_type,omitempty"`
	User      string    `json:"name,omitempty"`
	RoomID    int       `json:"room_id,omitempty"`
	Color     string    `json:"color,omitempty"`
	Msg       string    `json:"msg,omitempty"`
	Password  string    `json:"secret,omitempty"`
	Timestamp time.Time `json:"time,omitempty"`
}
