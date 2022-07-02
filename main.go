package main

import (
	"api_chat/server"
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	address := server.Config.Address
	// If port env var is set, PaaS platform (Heroku) is being used
	if port, ok := os.LookupEnv("PORT"); ok {
		address = "0.0.0.0:" + port
	}

	// starting up the server
	server := &http.Server{
		Addr:           address,
		Handler:        server.Mux,
		ReadTimeout:    time.Duration(server.Config.ReadTimeout * int64(time.Second)),
		WriteTimeout:   time.Duration(server.Config.WriteTimeout * int64(time.Second)),
		MaxHeaderBytes: 1 << 20,
	}
	fmt.Println("NEO-CHAT", version(), "started at", server.Addr)
	if _, exist := os.LookupEnv("PORT"); exist {
		// TLS is already enabled on Heroku PaaS platform
		if err := server.ListenAndServe(); err != nil {
			fmt.Println("Error starting server", err.Error())
		}
	} else {
		if err := server.ListenAndServeTLS("gencert/cert.pem", "gencert/key.pem"); err != nil {
			// If TLS fails e.g. because certs are missing on CI test env, we will fallback to regular HTTP
			if err := server.ListenAndServe(); err != nil {
				fmt.Println("Error starting server", err.Error())
			}
		}
	}
}

// version
func version() string {
	return "0.1"
}
