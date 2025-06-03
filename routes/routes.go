package routes

import (
	"html/template"
	"net/http"
	"tersoh-backend/controllers"
	"tersoh-backend/middleware"

	"github.com/gorilla/mux"
)

func SetupRouter() *mux.Router {
	r := mux.NewRouter()

	// Template rendering
	tmpl := template.Must(template.ParseGlob("./web/templates/*.tmpl"))

	// Serve static assets
	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets/"))))

	// CORS
	r.Use(middleware.Recover)
	r.Use(middleware.EnforceHTTPS)
	r.Use(middleware.CORS)

	api := r.PathPrefix("/api/v1").Subrouter()

	// Protected API endpoints
	api.Use(middleware.JWTMiddleware)
	api.HandleFunc("/posts", controllers.CreatePost).Methods("POST")
	api.HandleFunc("/posts", controllers.ListPosts).Methods("GET")
	api.HandleFunc("/messages", controllers.SendMessage).Methods("POST")
	api.HandleFunc("/messages", controllers.ListMessages).Methods("GET")
	api.HandleFunc("/transactions", controllers.CreateTransaction).Methods("POST")
	api.HandleFunc("/transactions", controllers.ListTransactions).Methods("GET")
	api.HandleFunc("/admin/analytics", controllers.ComputeAnalytics).Methods("GET")
	api.HandleFunc("/rates/average", controllers.GetAverageRates).Methods("GET")

	// User management API
	api.HandleFunc("/users", controllers.ListUsers).Methods("GET")
	api.HandleFunc("/users", controllers.CreateUser).Methods("POST")

	// Dynamic page routes
	pages := []string{
		"dashboard",
		"users",
		"userman",
		"userKYC",
		"chat",
		"changepass",
		"transaction",
		"dealprompt",
	}
	for _, p := range pages {
		p := p
		r.HandleFunc("/"+p, func(w http.ResponseWriter, r *http.Request) {
			tmpl.ExecuteTemplate(w, p+".tmpl", nil)
		}).Methods("GET")
	}

	return r
}
