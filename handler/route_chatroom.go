package handler

import (
	"api_chat/config"
	"api_chat/features"
	"api_chat/models"
	"api_chat/repository"
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
)

// HandleRoom main handler function
func HandleRoom(w http.ResponseWriter, r *http.Request) (err error) {
	queries := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")
	if titleOrID, ok := queries["titleOrID"]; ok {
		cr, err := repository.CS.Retrieve(titleOrID)
		if err != nil {
			config.Info("erroneous chats API request", r, err)
			return err
		}
		switch r.Method {
		case "GET":
			err = handleGet(w, cr)
			return err
		case "PUT":
			err = handlePut(w, r, cr, titleOrID)
			return err
		case "DELETE":
			err = handleDelete(w, cr)
			return err
		}
	} else {
		err = &config.APIError{
			Code: 103,
		}
	}

	return err
}

// Retrieve a chat room
// GET /chat/1
func handleGet(w http.ResponseWriter, cr *models.ChatRoom) (err error) {
	res, err := features.ToJSON(*cr)
	if err != nil {
		return
	}
	config.Info("retrieved chat room:", cr.Title)
	if _, err := w.Write(res); err != nil {
		config.Danger("Error writing", res)
	}
	return
}

// HandlePost Create a ChatRoom
// POST /chats
func HandlePost(w http.ResponseWriter, r *http.Request) (err error) {
	w.Header().Set("Content-Type", "application/json")
	// read in request
	contentLength := r.ContentLength
	body := make([]byte, contentLength)
	if _, err := r.Body.Read(body); err != nil {
		config.Danger("Error reading", r, err.Error())
	}
	// create ChatRoom obj
	var cr models.ChatRoom
	if err = json.Unmarshal(body, &cr); err != nil {
		config.Warning("error encountered reading POST:", err.Error())
		return err
	}
	if err = repository.CS.Add(&cr); err != nil {
		config.Warning("error encountered adding chat room:", err.Error())
		return err
	}
	// Retrieve updated object
	createdChatRoom, err := repository.CS.Retrieve(cr.Title)
	if err != nil {
		return err
	}
	res, _ := features.ToJSON(*createdChatRoom)
	w.WriteHeader(201)
	if _, err := w.Write(res); err != nil {
		config.Danger("Error writing", res)
	}
	return
}

// Update a room
// PUT /chats/<id>
func handlePut(w http.ResponseWriter, r *http.Request, currentChatRoom *models.ChatRoom, title string) (err error) {
	var cr models.ChatRoom
	contentLength := r.ContentLength
	body := make([]byte, contentLength)
	if _, err := r.Body.Read(body); err != nil {
		config.Danger("Error reading", r, err.Error())
	}
	if err = json.Unmarshal(body, &cr); err != nil {
		config.Warning("error encountered updating chat room:", err.Error())
		return
	}
	if err = repository.CS.Update(title, &cr); err != nil {
		config.Warning("error encountered updating chat room:", cr, err.Error())
		return
	}
	// Retrieve updated object
	modifiedChatRoom, err := repository.CS.RetrieveID(currentChatRoom.ID)
	if err != nil {
		return err
	}
	config.Info("updated chat room:", title)
	res, _ := features.ToJSON(*modifiedChatRoom)
	if _, err := w.Write(res); err != nil {
		config.Danger("Error writing", res)
	}
	return
}

// Delete a room
// DELETE /chat/<id>
func handleDelete(w http.ResponseWriter, cr *models.ChatRoom) (err error) {
	err = repository.CS.Delete(cr)
	if err != nil {
		config.Warning("error encountered deleting chat room:", err.Error())
		return
	}
	// report on status
	config.Info("deleted chat room:", cr.Title)
	config.ReportStatus(w, true, nil)
	return
}
