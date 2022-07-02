package server

import (
	"api_chat/config"
	"api_chat/handler"
	"api_chat/repository"
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
	//REST-API for chat room [JSON]
	api.Handle("/chats", handler.ErrHandler(handler.HandlePost)).Methods(http.MethodPost)
	api.Handle("/chats/{titleOrID}", handler.ErrHandler(handler.Authorize(handler.HandleRoom))).Methods(http.MethodGet, http.MethodPut, http.MethodDelete)
	// Check password matches room
	api.Handle("/chats/{titleOrID}/token", handler.ErrHandler(handler.Login)).Methods(http.MethodPost)
	// Check password matches room
	api.Handle("/chats/{titleOrID}/token/renew", handler.ErrHandler(handler.RenewToken)).Methods(http.MethodGet)
	// Chat Sessions (WebSocket)
	// Do not authorize since you can't add headers to WebSockets. We will do authorization when actually receiving chat messages
	api.Handle("/chats/{titleOrID}/ws", handler.Authorize(handler.WebSocketHandler)).Methods(http.MethodGet)
	return api
}

func init() {
	loadConfig()
	loadEnvs()
	loadLog()
	// initialize chat server
	repository.CS.Init()
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
		handler.SecretKey = key
	}
}
