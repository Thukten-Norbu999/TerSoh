package controllers

import (
	"encoding/json"
	"net/http"
	"time"

	"tersoh-backend/internal/utils"
	"tersoh-backend/models"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte("your_secret_key")

// Signup registers a new user
func Signup(w http.ResponseWriter, r *http.Request) {
	var creds struct{ Username, Password string }
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid payload")
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	user := models.User{Username: creds.Username, PasswordHash: string(hash)}
	if err := models.DB.Create(&user).Error; err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// Login authenticates and returns a JWT
func Login(w http.ResponseWriter, r *http.Request) {
	var creds struct{ Username, Password string }
	json.NewDecoder(r.Body).Decode(&creds)
	var user models.User
	if err := models.DB.Where("username = ?", creds.Username).First(&user).Error; err != nil {
		utils.RespondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(creds.Password)) != nil {
		utils.RespondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}
	expiration := time.Now().Add(24 * time.Hour)
	claims := models.Claims{Username: creds.Username, StandardClaims: jwt.StandardClaims{ExpiresAt: expiration.Unix()}}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ts, _ := token.SignedString(jwtKey)
	// record event
	models.DB.Create(&models.LoginEvent{Username: creds.Username, Timestamp: time.Now()})
	utils.RespondJSON(w, http.StatusOK, map[string]string{"token": ts})
}

// ChangePassword allows users to change password
func ChangePassword(w http.ResponseWriter, r *http.Request) {
	// TODO: implement JWT middleware to set user
	var pw struct{ Old, New string }
	json.NewDecoder(r.Body).Decode(&pw)
	// Placeholder: accept and return OK
	w.WriteHeader(http.StatusOK)
}
