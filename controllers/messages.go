package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"tersoh-backend/internal/utils"
	"tersoh-backend/models"
)

func SendMessage(w http.ResponseWriter, r *http.Request) {
	var m models.Message
	json.NewDecoder(r.Body).Decode(&m)
	models.DB.Create(&m)
	w.WriteHeader(http.StatusCreated)
}

func GetMessages(w http.ResponseWriter, r *http.Request) {
	var msgs []models.Message
	models.DB.Find(&msgs)
	utils.RespondJSON(w, http.StatusOK, msgs)
}

// ListMessages returns paginated list of messages
func ListMessages(w http.ResponseWriter, r *http.Request) {
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

	var msgs []models.Message
	models.DB.
		Order("created_at desc").
		Limit(limit).
		Offset(offset).
		Find(&msgs)
	utils.RespondJSON(w, http.StatusOK, msgs)
}
