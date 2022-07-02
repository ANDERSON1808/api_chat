package handler

import (
	"api_chat/config"
	"api_chat/features"
	"api_chat/models"
	"api_chat/repository"
	"encoding/json"
	"net/http"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/DavidSchott/chitchat/data"
	"github.com/gorilla/mux"
)

var SecretKey string = "my_secret_random_key_>_than_24_characters"

// Add authorization
// POST /chats/{titleOrID}/token
func Login(w http.ResponseWriter, r *http.Request) (err error) {
	w.Header().Set("Content-Type", "application/json")
	// read in request
	len := r.ContentLength
	body := make([]byte, len)
	if _, err := r.Body.Read(body); err != nil {
		config.Danger("Error reading request", r)
	}
	var c models.ChatEvent
	if err := json.Unmarshal(body, &c); err != nil {
		config.Danger("Error parsing token request", r)
	}
	queries := mux.Vars(r)
	if titleOrID, ok := queries["titleOrID"]; ok {
		cr, err := repository.CS.Retrieve(titleOrID)
		if err != nil {
			config.Info("erroneous chats API request", r, err)
			return err
		}
		if cr.Type == models.PublicRoom {
			// Ignore public room
			config.ReportStatus(w, true, nil)
		} else if features.MatchesPassword(c.Password, *cr) {
			if c.User == "" {
				return &config.APIError{
					Code:  303,
					Field: "name",
				}
			}
			// Success! Generate token using secret key concatenated with room's password (length > 32)
			tokenString, err := config.EncodeJWT(&c, cr, generateUniqueKey(cr))
			if err != nil {
				return err
			}
			// Success, respond with token in JSON body
			jsonEncoding, _ := json.Marshal(struct {
				Outcome  bool   `json:"status"`
				Username string `json:"name"`
				RoomID   int    `json:"room_id"`
				Token    string `json:"token"`
			}{
				Outcome:  true,
				Username: c.User,
				RoomID:   cr.ID,
				Token:    tokenString,
			})
			w.WriteHeader(http.StatusCreated)
			if _, err := w.Write(jsonEncoding); err != nil {
				config.Danger("Error writing", jsonEncoding)
			}

		} else {
			return &config.APIError{
				Code:  304,
				Field: "secret",
			}
		}
	}
	return
}

// RenewToken Refreshes tokens before they expire
// GET /chats/{titleOrID}/token/renew
func RenewToken(w http.ResponseWriter, r *http.Request) (err error) {
	w.Header().Set("Content-Type", "application/json")
	queries := mux.Vars(r)
	if titleOrID, ok := queries["titleOrID"]; ok {
		cr, err := repository.CS.Retrieve(titleOrID)
		if err != nil {
			config.Info("erroneous chats API request", r, err)
			return err
		}
		if cr.Type == models.PublicRoom {
			// Ignore public room
			config.ReportStatus(w, true, nil)
		} else {
			// Check authorization header
			// Get the JWT string from the header
			tknStr, err := extractJwtToken(r)
			if err != nil {
				return &config.APIError{
					Code:  403,
					Field: "token",
				}
			}
			claim := &config.Claims{}
			if err = config.ParseJWT(tknStr, claim, generateUniqueKey(cr)); err != nil {
				return err
			}
			// Success! Generate token
			tokenStringNew, err := claim.RefreshJWT(generateUniqueKey(cr))
			if err != nil {
				return err
			}
			// Success, respond with token in JSON body
			jsonEncoding, _ := json.Marshal(struct {
				Outcome  bool   `json:"status"`
				Username string `json:"name"`
				RoomID   int    `json:"room_id"`
				Token    string `json:"token"`
			}{
				Outcome:  true,
				Username: claim.Username,
				RoomID:   cr.ID,
				Token:    tokenStringNew,
			})
			w.WriteHeader(http.StatusCreated)
			if _, err := w.Write(jsonEncoding); err != nil {
				config.Danger("Error writing", jsonEncoding)
			}
		}
	}
	return
}

// Authorize will call the handler if authorization bearer token is valid. Otherwise, it will send a failed outcome
func Authorize(h ErrHandler) ErrHandler {
	return func(w http.ResponseWriter, r *http.Request) (err error) {
		// Skip authorization for special case of GET /chats/<id> for now
		// TODO: Rewrite client-side app to request token before GET chat room
		if name := runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name(); strings.HasSuffix(name, "handleRoom") && r.Method == http.MethodGet {
			return h(w, r)
		}
		queries := mux.Vars(r)
		if titleOrID, ok := queries["titleOrID"]; ok {
			cr, err := repository.CS.Retrieve(titleOrID)
			if err != nil {
				config.Info("erroneous chats API request", r, err)
				return err
			}
			if cr.Type != models.PublicRoom {
				// Check authorization header
				// Get the JWT string from the cookie
				tknStr, err := extractJwtToken(r)
				if err != nil {
					return &config.APIError{
						Code:  403,
						Field: "token",
					}
				}
				claim := &config.Claims{}
				err = config.ParseJWT(tknStr, claim, generateUniqueKey(cr))
				if err != nil {
					return err
				}
			}

			// Success, call h(w,r)
			return h(w, r)
		}
		return
	}
}

// Strips 'Token' or 'Bearer' prefix from token string
func stripTokenPrefix(tok string) string {
	// split token to 2 parts
	tokenParts := strings.Split(tok, " ")
	if len(tokenParts) < 2 {
		return tokenParts[0]
	}
	return tokenParts[1]
}

// extractJwtToken extracts token from Authorization header
func extractJwtToken(req *http.Request) (string, error) {
	// Strip "Bearer" from Authorization: Bearer <token>
	tokenString := stripTokenPrefix(req.Header.Get("Authorization"))
	if tokenString == "" {
		// Want to check
		tokenString = req.Header.Get("Sec-WebSocket-Protocol")
	}

	if tokenString == "" {
		return "", &data.APIError{Code: 403, Field: "token"}
	}

	return tokenString, nil
}

// Generate unique key should ensure that the generated key is unique for a given room
// This key does not need to be unique per user necessarily since the token will be unique
func generateUniqueKey(cr *models.ChatRoom) string {
	return SecretKey + cr.Password + strconv.Itoa(cr.ID)
}
