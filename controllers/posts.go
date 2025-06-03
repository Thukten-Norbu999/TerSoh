package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"tersoh-backend/internal/utils"
	"tersoh-backend/models"

	"github.com/gorilla/mux"
)

func CreatePost(w http.ResponseWriter, r *http.Request) {
	var p models.Post
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid payload")
		return
	}
	models.DB.Create(&p)
	utils.RespondJSON(w, http.StatusCreated, p)
}

func GetPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	var p models.Post
	if err := models.DB.First(&p, id).Error; err != nil {
		utils.RespondError(w, http.StatusNotFound, "Post not found")
		return
	}
	utils.RespondJSON(w, http.StatusOK, p)
}

// ListPosts returns paginated list of posts
func ListPosts(w http.ResponseWriter, r *http.Request) {
	// parse pagination params
	page := 1
	limit := 10
	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	offset := (page - 1) * limit

	var posts []models.Post
	models.DB.
		Order("created_at desc").
		Limit(limit).
		Offset(offset).
		Find(&posts)
	utils.RespondJSON(w, http.StatusOK, posts)
}
