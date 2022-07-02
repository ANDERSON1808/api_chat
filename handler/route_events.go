package handler

import (
	"api_chat/config"
	"api_chat/features"
	"api_chat/models"
	"api_chat/repository"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var (
	upgrade = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow connections from any origin.
		},
	}
)

// WebSocketHandler Upgrade to a ws connection
// Add to active chat session
// GET /chats/{titleOrID}/ws
func WebSocketHandler(w http.ResponseWriter, r *http.Request) (err error) {
	queries := mux.Vars(r)
	if titleOrID, ok := queries["titleOrID"]; ok {
		// Fetch room & authorize
		cr, err := repository.CS.Retrieve(titleOrID)
		if err != nil {
			config.Warning("Error retrieving room", r, err)
			return err
		}
		// Do stuff here:
		wsConn, err := upgrade.Upgrade(w, r, nil)
		if err != nil {
			errorMessage(w, r, "Critical error creating WebSocket: "+err.Error())
			config.Danger("error creating WebSocket: ", err)
			return &config.APIError{Code: 301}
		}
		client := &models.Client{Room: cr, Conn: wsConn, Send: make(chan []byte)}
		client.Room.Broker.OpenClient <- client

		// Allow collection of memory referenced by the caller by doing all work in
		// new goroutines.
		go features.WritePump(client)
		go features.ReadPump(client)
	} else {
		return &config.APIError{Code: 101}
	}

	return
}
