package features

import (
	"api_chat/config"
	"api_chat/models"
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"strings"
)

// ToJSON marshals a ChatRoom object in a JSON encoding that can be returned to users
func ToJSON(cr models.ChatRoom) (jsonEncoding []byte, err error) {
	// Populate client slice. TODO: Can this be simplified?
	clientsSlice := make([]models.Client, len(cr.Clients))
	var i int = 0
	for _, v := range cr.Clients {
		//clientsSlice = append(clientsSlice, *v)
		clientsSlice[i] = *v
		i++
	}
	// Create new JSON struct with clients
	jsonEncoding, err = json.Marshal(struct {
		*models.ChatRoom
		Clients []models.Client `json:"users"`
	}{
		ChatRoom: &cr,
		Clients:  clientsSlice,
	})
	return jsonEncoding, err
}

//AddClient will add a user to a ChatRoom
func AddClient(c *models.Client, cr models.ChatRoom) (err error) {
	if clientExists(c.Username, cr) {
		return &config.APIError{
			Code:  202,
			Field: c.Username,
		}
	}
	cr.Clients[strings.ToLower(c.Username)] = c
	return
}

// RemoveClient will remove a user from a ChatRoom
func RemoveClient(user string, cr models.ChatRoom) (err error) {
	if !clientExists(user, cr) {
		return &config.APIError{
			Code:  201,
			Field: user,
		}
	}
	delete(cr.Clients, strings.ToLower(user))
	return
}

// Authorize authorizes a given ChatEvent for the Room
func Authorize(c *models.ChatEvent, cr models.ChatRoom) bool {
	return MatchesPassword(c.Password, cr)
}

// IsValid validates a chat room fields are still valid
func IsValid(cr models.ChatRoom) (err *config.APIError, validity bool) {
	// Title should be at least 2 characters
	if len(cr.Title) < 2 || len(cr.Title) > 70 {
		return &config.APIError{
			Code:  105,
			Field: "title",
		}, false
	}
	// Description shall not be too long
	if len(cr.Description) > 70 {
		return &config.APIError{
			Code:  105,
			Field: "description",
		}, false
	}
	visibility := strings.ToLower(cr.Type)
	// Visibility must be set
	if visibility != models.PublicRoom && visibility != models.PrivateRoom && visibility != models.HiddenRoom {
		return &config.APIError{
			Code:  105,
			Field: "visibility",
		}, false
	}
	// Non-public rooms require a valid password
	if (len(cr.Password) < 8) && visibility != models.PublicRoom {
		return &config.APIError{
			Code:  105,
			Field: "password",
		}, false
	}
	// A public room should not have a password set (to avoid accidents)
	if len(cr.Password) != 0 && visibility == models.PublicRoom {
		return &config.APIError{
			Code:  105,
			Field: "visibility",
		}, false
	}
	return nil, true
}

// MatchesPassword takes in a value and compares it with the room's password
func MatchesPassword(val string, cr models.ChatRoom) bool {
	err := bcrypt.CompareHashAndPassword([]byte(cr.Password), []byte(val))
	return err == nil
}

func clientExists(name string, cr models.ChatRoom) bool {
	name = strings.ToLower(name)
	for k := range cr.Clients {
		if k == name {
			return true
		}
	}
	return false
}

// PrettyTime prints the creation date in a pretty format
func PrettyTime(cr models.ChatRoom) string {
	layout := "Mon Jan _2 15:04"
	return cr.CreatedAt.Format(layout)
}

// Participants prints the # of active clients
func Participants(cr models.ChatRoom) int {
	return len(cr.Clients)
}
