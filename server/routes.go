package handler

import (
	"api_chat/config"
	"api_chat/server"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

// Configuration stores config info of server
type Configuration struct {
	Address      string
	RedisURL     string
	ReadTimeout  int64
	WriteTimeout int64
	Static       string
}

// Config captures parsed input from config.json
var Config Configuration

// Mux contains all the HTTP handlers
var (
	Mux *mux.Router
)

// registerHandlers will register all HTTP handlers
func registerHandlers() *mux.Router {
	api := mux.NewRouter()
	// handle static assets by routing requests from /static/ => "public" directory
	staticDir := "/static/"
	api.PathPrefix(staticDir).Handler(http.StripPrefix(staticDir, http.FileServer(http.Dir(Config.Static))))
	//REST-API for chat room [JSON]
	api.Handle("/chats", ErrHandler(HandlePost)).Methods(http.MethodPost)
	api.Handle("/chats/{titleOrID}", ErrHandler(Authorize(HandleRoom))).Methods(http.MethodGet, http.MethodPut, http.MethodDelete)
	// Check password matches room
	api.Handle("/chats/{titleOrID}/token", ErrHandler(Login)).Methods(http.MethodPost)
	// Check password matches room
	api.Handle("/chats/{titleOrID}/token/renew", ErrHandler(RenewToken)).Methods(http.MethodGet)
	// Chat Sessions (WebSocket)
	// Do not authorize since you can't add headers to WebSockets. We will do authorization when actually receiving chat messages
	api.Handle("/chats/{titleOrID}/ws", ErrHandler(Authorize(WebSocketHandler))).Methods(http.MethodGet)
	api.HandleFunc("/favicon.ico", faviconHandler)
	return api
}

func init() {
	loadConfig()
	loadEnvs()
	loadLog()
	// initialize chat server
	server.CS.Init()
	Mux = registerHandlers()
}

func loadLog() {
	file, err := os.OpenFile("chitchat.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		file, err = os.OpenFile("../chitchat.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalln("Failed to open log file", err)
		}

	}
	config.Logger = log.New(file, "INFO ", log.Ldate|log.Ltime|log.Lshortfile)
}

func loadConfig() {
	file, err := os.Open("config.json")
	if err != nil {
		file, err = os.Open("../config.json")
		if err != nil {
			log.Fatalln("Cannot open config file", err)
		}
	}
	decoder := json.NewDecoder(file)
	Config = Configuration{}
	err = decoder.Decode(&Config)
	if err != nil {
		log.Fatalln("Cannot get configuration from file", err)
	}
}

func loadEnvs() {
	if key, ok := os.LookupEnv("SECRET_KEY"); ok {
		SecretKey = key
	}
}
func faviconHandler(w http.ResponseWriter, r *http.Request) {
	//w.Header().Set("Content-Type", "image/x-icon")
	//w.Header().Set("Cache-Control", "public, max-age=7776000")
	http.ServeFile(w, r, "public/img/favicon.ico")
}
