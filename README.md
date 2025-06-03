[ ] models.DB.create -> issues in the controllers

[ ] Make sure you go through all the issues that comes with scrape.go(mainly playwright)

[ ] Go through the .env and the main required are forex, kyc_client, secret_key(replace)

[ ] Use the following for internal\thirdparty\kyc
```
package thirdparty

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"your-module-path/model" // Replace this with your actual module path
)

const BaseUrl string = "https://apx.didit.me"

type TokenResponse struct {
	Token      string `json:"access_token"`
	Expires_in int    `json:"expires_in"`
}

func KYCAccessToken() ([]byte, error) {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file:", err)
		return nil, err
	}

	clientID := os.Getenv("KYC_CLIENT_ID")
	clientSK := os.Getenv("KYC_CLIENT_SK")
	if clientID == "" || clientSK == "" {
		return nil, fmt.Errorf("client ID or secret key missing in environment variables")
	}

	encoded := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSK))
	authHeader := fmt.Sprintf("Basic %s", encoded)

	params := url.Values{}
	params.Add("grant_type", "client_credentials")

	apiURL := BaseUrl + "/auth/v2/token/"
	req, err := http.NewRequest("POST", apiURL, bytes.NewBufferString(params.Encode()))
	if err != nil {
		log.Println("Failed to create request:", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", authHeader)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Request failed:", err)
		return nil, err
	}
	defer resp.Body.Close()

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Failed to read response:", err)
		return nil, err
	}

	return responseData, nil
}

func RetrieveKycUser(sessID string) ([]byte, error) {
	if sessID == "" {
		return nil, fmt.Errorf("empty session ID")
	}

	url := fmt.Sprintf("https://verification.didit.me/v1/session/%s/decision", sessID)
	token, err := KYCAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token")
	}

	var tokenData TokenResponse
	if err := json.Unmarshal(token, &tokenData); err != nil {
		return nil, fmt.Errorf("invalid token format")
	}
	if tokenData.Token == "" {
		return nil, fmt.Errorf("empty access token received")
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenData.Token))

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response")
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var jsonCheck map[string]interface{}
		if err := json.Unmarshal(body, &jsonCheck); err != nil {
			return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, http.StatusText(resp.StatusCode))
		}
		var apiError struct {
			Message string `json:"message"`
		}
		if err := json.Unmarshal(body, &apiError); err == nil && apiError.Message != "" {
			return nil, fmt.Errorf("API error: %s", apiError.Message)
		}
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var jsonResponse map[string]interface{}
	if err := json.Unmarshal(body, &jsonResponse); err != nil {
		return nil, fmt.Errorf("invalid response format")
	}

	return body, nil
}

type KycHandler struct {
	DB *gorm.DB
}

func (k *KycHandler) RetrieveKycUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionid"]

	resp, err := RetrieveKycUser(sessionID)
	if err != nil {
		if err.Error() == "API returned status 404" {
			var verification model.Verification
			result := k.DB.Where("session_id=?", sessionID).First(&verification)
			if result.Error != nil {
				log.Printf("Session ID %s not found in DB: %v", sessionID, result.Error)
			}
			deleteResult := k.DB.Delete(&model.Verification{}, "session_id = ?", sessionID)
			if deleteResult.Error != nil {
				log.Printf("Failed to delete session ID %s: %v", sessionID, deleteResult.Error)
			} else {
				log.Printf("Deleted session ID %s from database", sessionID)
			}
		}

		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": err.Error(),
			"type":    "error",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "successful",
		"type":    "success",
		"data":    json.RawMessage(resp),
	})
}

func StartServer(db *gorm.DB) {
	r := mux.NewRouter()
	k := &KycHandler{DB: db}

	r.HandleFunc("/kyc/session/{sessionid}", k.RetrieveKycUserHandler).Methods("GET")

	srv := &http.Server{
		Handler:      r,
		Addr:         ":8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Println("Server starting on http://localhost:8080")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

```

[ ] 