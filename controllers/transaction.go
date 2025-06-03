package controllers

import (
	"encoding/json"
	"net/http"

	"tersoh-backend/models"
)

func CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var t models.Transaction
	json.NewDecoder(r.Body).Decode(&t)
	models.DB.Create(&t)
	w.WriteHeader(http.StatusCreated)
}
