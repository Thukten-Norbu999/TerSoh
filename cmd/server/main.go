package main

import (
	log "github.com/sirupsen/logrus"

	"net/http"
	"tersoh-backend/config"
	"tersoh-backend/routes"
)

func main() {
	// Structured logging setup
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)

	config.InitDB()
	router := routes.SetupRouter()
	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
