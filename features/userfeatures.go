package features

import (
	"api_chat/models"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"time"
)

// ReadPump pumps messages from the websocket connection to the broker.
//
// The application runs ReadPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func ReadPump(c *models.Client) {
	defer func() {
		c.Room.Broker.CloseClient <- c
		err := c.Conn.Close()
		if err != nil {
			return
		}
	}()
	c.Conn.SetReadLimit(models.MaxMessageSize)
	if err := c.Conn.SetReadDeadline(time.Now().Add(models.PongWait)); err != nil {
		log.Println("Error setting pongWait read deadline", err.Error())
	}
	c.Conn.SetPongHandler(func(string) error {
		if err := c.Conn.SetReadDeadline(time.Now().Add(models.PongWait)); err != nil {
			log.Println("Error setting pongWait read deadline", err.Error())
		}
		return nil
	})
	for {
		mt, data, err := c.Conn.ReadMessage() // TODO: Switch to ReadJSON
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) || err == io.EOF {
				res, _ := json.Marshal(&models.ChatEvent{User: c.Username, Msg: fmt.Sprintf("%s has left the room.", c.Username), Color: c.Color})
				c.Room.Broker.Notification <- res
			}
			unsubscribe(&models.ChatEvent{User: c.Username, Color: c.Color}, c)
			log.Printf("error: %v", err)
			break
		}
		switch mt {
		case websocket.TextMessage:
			ce, err := ValidateEvent(data)
			if err != nil {
				log.Printf("Error parsing JSON ChatEvent: %v", err)
				break
			}
			// Set timestamp and room ID
			ce.Timestamp = time.Now()
			ce.RoomID = c.Room.ID

			// Perform requested action
			switch ce.EventType {
			case models.Unsubscribe:
				// Populate activity
				c.Room.Clients[ce.User].LastActivity = ce.Timestamp
				unsubscribe(&ce, c)
			case models.Subscribe:
				// LastActivity will be populated in subscribe
				subscribe(&ce, c)
			case models.Broadcast:
				// Populate activity
				c.Room.Clients[ce.User].LastActivity = ce.Timestamp
				broadcast(&ce, c)
			default:
				// Populate activity
				//c.Room.Clients[ce.User].LastActivity = ce.Timestamp
				//broadcast(&ce,c)
				log.Printf("Warning: unknown event type %s", ce.EventType)
			}

		default:
			log.Printf("Warning: unknown message type")
		}
	}
}

// WritePump pumps messages from the broker to the websocket connection.
//
// A goroutine running WritePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func WritePump(c *models.Client) {
	ticker := time.NewTicker(models.PingPeriod)
	defer func() {
		ticker.Stop()
		err := c.Conn.Close()
		if err != nil {
			return
		}
	}()
	for {
		select {
		case message, ok := <-c.Send:
			if err := c.Conn.SetWriteDeadline(time.Now().Add(models.WriteWait)); err != nil {
				log.Println("Error setting writeWait write deadline", err.Error())
			}
			if !ok {
				// The broker closed the channel.
				if err := c.Conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					log.Println("Error writing WebSocket closing message:", err.Error())
				}
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			if _, err := w.Write(message); err != nil {
				log.Printf("Error writing message. Error: %s", err.Error())
			}

			// Add queued chat messages to the current websocket message.
			n := len(c.Send)
			for i := 0; i < n; i++ {
				if _, err := w.Write(<-c.Send); err != nil {
					log.Printf("Error writing sent message. Error: %s", err.Error())
				}
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.Conn.SetWriteDeadline(time.Now().Add(models.WriteWait)); err != nil {
				log.Println("Error setting writeWait write deadline", err.Error())
			}
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func formatEventData(c *models.ChatEvent) []byte {
	data, _ := json.Marshal(c)
	return data
}

func broadcast(evt *models.ChatEvent, c *models.Client) {
	evt.EventType = models.Broadcast
	c.Room.Broker.Notification <- formatEventData(evt)
}

func subscribe(evt *models.ChatEvent, c *models.Client) {
	// Init client values
	c.Username = evt.User
	c.Color = evt.Color
	c.LastActivity = time.Now()
	if err := AddClient(c, *c.Room); err != nil {
		log.Println("error adding client:", err.Error())
		return
	}
	log.Println("Adding client to Chatroom: ", evt.User)
	evt.EventType = models.Subscribe
	evt.Msg = fmt.Sprintf("%s entered the room.", evt.User)
	go func() {
		time.Sleep(200 * time.Millisecond)
		c.Room.Broker.Notification <- formatEventData(evt)
	}()
}

func unsubscribe(evt *models.ChatEvent, c *models.Client) {
	// Remove Client from tracked list
	if err := RemoveClient(evt.User, *c.Room); err != nil {
		log.Println("Error removing client", err.Error())
	}
	log.Println(fmt.Sprintf("Unsubscribing %s in room %d", evt.User, c.Room.ID))
	evt.EventType = models.Unsubscribe
	evt.Msg = fmt.Sprintf("%s has left the room.", evt.User)
	go func() {
		time.Sleep(200 * time.Millisecond)
		c.Room.Broker.Notification <- formatEventData(evt)
	}()
}
