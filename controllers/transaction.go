package controllers

import (
	"encoding/json"
	"net/http"

	"tersoh-backend/config"
	"tersoh-backend/models"
	"tersoh-backend/utils"
)

func CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var t models.Transaction
	json.NewDecoder(r.Body).Decode(&t)
	models.DB.Create(&t)
	w.WriteHeader(http.StatusCreated)
}

func ListTransactions(w http.ResponseWriter, r *http.Request) {
	var ts []models.Transaction
	config.DB.Order("created_at desc").Find(&ts)
	utils.RespondJSON(w, http.StatusOK, ts)
}
