package repository

import (
	"api_chat/config"
	"api_chat/features"
	"api_chat/models"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// ChatServer maintains all ChatRooms. TODO: This will be replaced by a database soon
type ChatServer struct {
	RoomsID map[int]*models.ChatRoom
	Rooms   map[string]*models.ChatRoom // TODO: Remove this duplication once data layer moves to DB
	Index   *int
}

var index int

// CS is the global ChatServer referencing all chat room objects
var CS ChatServer = ChatServer{
	RoomsID: make(map[int]*models.ChatRoom),
	Rooms:   make(map[string]*models.ChatRoom),
	Index:   &index,
}

// Init will initialize the ChatServer with the default public room.
func (cs ChatServer) Init() {
	CS.push(&models.ChatRoom{
		Title:       "Public Chat",
		Description: "This is the default chat, available to everyone!",
		Type:        "public",
		Password:    "",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		ID:          1,
		Broker:      models.NewBroker(1),
		Clients:     make(map[string]*models.Client),
	})
}

func (cs ChatServer) push(cr *models.ChatRoom) {
	// Update indices, create new session
	*cs.Index++
	// TODO: Generate UUIDs?
	cr.ID = *cs.Index
	cr.Clients = make(map[string]*models.Client)
	cr.Type = strings.ToLower(cr.Type)
	cr.Broker = models.NewBroker(cr.ID)
	// Start broker for rooms
	go cr.Broker.Listen()
	// Push to chat server
	cs.Rooms[strings.ToLower(cr.Title)] = cr
	cs.RoomsID[cr.ID] = cr
}

func (cs ChatServer) pop(title string, ID int) {
	delete(cs.Rooms, strings.ToLower(title))
	delete(cs.RoomsID, ID)
	*cs.Index--
}

// Chats will return all non-hidden ChatRooms
func (cs ChatServer) Chats() (rooms []models.ChatRoom, err error) {
	rooms = make([]models.ChatRoom, 0)
	for _, v := range CS.Rooms {
		if v.Type != models.HiddenRoom {
			rooms = append(rooms, *v)
		}
	}
	return
}

// Retrieve returns a single chat room based on title or ID
func (cs ChatServer) Retrieve(title string) (cr *models.ChatRoom, err error) {
	if !cs.roomExists(title) {
		return cr, &config.APIError{
			Code:  101,
			Field: title,
		}
	}
	if id := config.IsInt(title); id != -1 {
		cr = cs.RoomsID[id]
	} else {
		cr = cs.Rooms[strings.ToLower(title)]
	}
	//err = Db.QueryRow("select id, content, author from posts where id = $1", id).Scan(&post.Id, &post.Content, &post.Author)
	return cr, nil
}

// RetrieveID returns a single chat room based on ID. NOTE: This has no error handling unlike cs.Retrieve()
func (cs ChatServer) RetrieveID(ID int) (cr *models.ChatRoom, err error) {
	cr = cs.RoomsID[ID]
	//err = Db.QueryRow("select id, content, author from posts where id = $1", id).Scan(&post.Id, &post.Content, &post.Author)
	return
}

func (cs ChatServer) roomExists(titleorID string) bool {
	if id, err := strconv.Atoi(titleorID); err == nil {
		for k := range CS.RoomsID {
			if k == id {
				return true
			}
		}
	} else {
		titleorID = strings.ToLower(titleorID)
		for k := range CS.Rooms {
			if strings.ToLower(k) == titleorID {
				return true
			}
		}
	}
	return false
}

// Add will create a new chat room and add it to the server
func (cs ChatServer) Add(cr *models.ChatRoom) (err error) {
	// validate chat room request
	if rapier, valid := features.IsValid(*cr); !valid {
		return rapier
	}
	if cs.roomExists(cr.Title) { // TODO: What if the room is hidden? Return unspecified error or inform user?
		return &config.APIError{
			Code:  102,
			Field: "title",
		}
	}
	cr.Type = strings.ToLower(cr.Type)
	if cr.Type != models.PublicRoom {
		pass, err := bcrypt.GenerateFromPassword([]byte(cr.Password), bcrypt.DefaultCost)
		if err != nil {
			return &config.APIError{
				Code:  104,
				Field: "secret",
			}
		}
		cr.Password = string(pass)
	} else if cr.Type == models.PublicRoom {
		cr.Password = ""
	}

	cr.CreatedAt = time.Now()
	cr.UpdatedAt = time.Now()
	cs.push(cr)
	return
}

// Update a chat room. NOTE: Authorization should have been done before calling this
// TODO: Get input from requested ID. Edit both RoomsID and Rooms.
func (cs ChatServer) Update(titleOrID string, modifiedChatRoom *models.ChatRoom) (err error) {
	currentChatRoom, err := cs.Retrieve(titleOrID)
	if err != nil {
		return
	}
	// Update password for validation
	modifiedChatRoom.Password = currentChatRoom.Password
	if apierr, valid := features.IsValid(*modifiedChatRoom); !valid {
		return apierr
	}
	// Update chat room
	// TODO: Allow updating Password?
	modifiedChatRoom.ID = currentChatRoom.ID
	modifiedChatRoom.UpdatedAt = time.Now()
	*currentChatRoom = *modifiedChatRoom
	//_, err = Db.Exec("update posts set content = $2, author = $3 where id = $1", post.Id, post.Content, post.Author)
	return
}

// Delete a chat room
func (cs ChatServer) Delete(cr *models.ChatRoom) (err error) {
	cs.pop(strings.ToLower(cr.Title), cr.ID)
	//_, err = Db.Exec("delete from posts where id = $1", post.Id)
	return
}
