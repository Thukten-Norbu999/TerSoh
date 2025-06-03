package routes

import (
	"net/http"
	"path/filepath"

	"tersoh-backend/controllers"

	"github.com/gorilla/mux"
)

func SetupRouter() *mux.Router {
	router := mux.NewRouter()

	api := router.PathPrefix("/api/v1").Subrouter()
	// Auth
	api.HandleFunc("/auth/signup", controllers.Signup).Methods("POST")
	api.HandleFunc("/auth/login", controllers.Login).Methods("POST")
	// Posts
	api.HandleFunc("/posts", controllers.CreatePost).Methods("POST")
	api.HandleFunc("/posts", controllers.GetPost).Methods("GET")
	// TODO: Add remaining controller routes here

	// Serve static assets
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(filepath.Join("web", "public", "static")))))

	return router
}
