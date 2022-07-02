package models

import (
	"time"
)

const (
	// PublicRoom is a room open for anyone to join without authentication
	PublicRoom = "public"
	// PrivateRoom is password protected and requires an authentication token in order to process requests
	PrivateRoom = "private"
	// HiddenRoom is a private room that is not listed on public-facing APIs. TODO: Hide this from GET /chats/<id> as well?
	HiddenRoom = "hidden"
)

// ChatRoom is a struct representing a chat room
// TODO:  Add Administrator
type ChatRoom struct {
	Title       string             `json:"title"`
	Description string             `json:"description,omitempty"`
	Type        string             `json:"visibility"`
	Password    string             `json:"password,omitempty"`
	CreatedAt   time.Time          `json:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt"`
	ID          int                `json:"id"`
	Broker      *Broker            `json:"-"`
	Clients     map[string]*Client `json:"-"`
}
