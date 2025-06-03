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

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

const BaseUrl string = "https://apx.didit.me"

type TokenResponse struct {
	Token      string `json:"access_token"`
	Expires_in int    `json:"expires_in"`
}

func KYCAccessToken() ([]byte, error) {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file:", err)
		return nil, err
	}

	// Get credentials from environment
	clientID := os.Getenv("KYC_CLIENT_ID")
	clientSK := os.Getenv("KYC_CLIENT_SK")

	if clientID == "" || clientSK == "" {
		return nil, fmt.Errorf("client ID or secret key missing in environment variables")
	}

	// Encode credentials in Base64 for Basic Auth
	encoded := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSK))
	authHeader := fmt.Sprintf("Basic %s", encoded)

	// Create URL-encoded form data
	params := url.Values{}
	params.Add("grant_type", "client_credentials")

	// Define API URL
	apiURL := BaseUrl + "/auth/v2/token/"

	req, err := http.NewRequest("POST", apiURL, bytes.NewBufferString(params.Encode()))
	if err != nil {
		log.Println("Failed to create request:", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", authHeader)

	// Send HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Request failed:", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Read and return raw response body (already JSON)
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Failed to read response:", err)
		return nil, err
	}

	return responseData, nil
}

type RequestOptions struct {
	Method  string
	Headers map[string]string
}

func RetrieveKycUser(sessID string) ([]byte, error) {
	// Validate session ID input
	if sessID == "" {
		return nil, fmt.Errorf("empty session ID")
	}

	url := fmt.Sprintf("https://verification.didit.me/v1/session/%s/decision", sessID)

	// Get access token
	token, err := KYCAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token")
	}

	// Parse token response
	var tokenData TokenResponse
	if err := json.Unmarshal(token, &tokenData); err != nil {
		return nil, fmt.Errorf("invalid token format")
	}
	if tokenData.Token == "" {
		return nil, fmt.Errorf("empty access token received")
	}

	// Create request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenData.Token))

	// Configure client with timeout
	client := &http.Client{Timeout: 30 * time.Second}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response")
	}

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Ensure the response is JSON, not HTML
		var jsonCheck map[string]interface{}
		if err := json.Unmarshal(body, &jsonCheck); err != nil {

			return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, http.StatusText(resp.StatusCode))
		}

		// Try to extract error message from response JSON
		var apiError struct {
			Message string `json:"message"`
		}
		if err := json.Unmarshal(body, &apiError); err == nil && apiError.Message != "" {
			return nil, fmt.Errorf("API error: %s", apiError.Message)
		}
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	// Validate that response is actually JSON
	var jsonResponse map[string]interface{}
	if err := json.Unmarshal(body, &jsonResponse); err != nil {
		return nil, fmt.Errorf("invalid response format")
	}

	return body, nil
}

type KycHandler struct {
	DB *gorm.DB
}

func (k *KycHandler) RetrieveKycUserHandler(c *gin.Context) {
	sessionID := c.Param("sessionid")

	// Call RetrieveKycUser function
	resp, err := RetrieveKycUser(sessionID)
	if err != nil {
		if err.Error() == "API returned status 404" {
			// Check if the session ID exists in the database
			var verification model.Verification
			result := k.DB.Where("session_id=?", sessionID).First(&verification)

			// Log the result of the session check
			if result.Error != nil {
				log.Printf("Session ID %s not found in the database: %v", sessionID, result.Error)
				// Return 404 because session doesn't exist

			}

			// Log that session ID exists before deletion

			// Attempt to delete the session record using Delete() method
			deleteResult := k.DB.Delete(&model.Verification{}, "session_id = ?", sessionID)

			// Check if deletion was successful and log the outcome
			if deleteResult.Error != nil {
				log.Printf("Failed to delete session ID %s: %v", sessionID, deleteResult.Error)

			}

			// Log successful deletion
			log.Printf("Session ID %s successfully deleted from the database", sessionID)
			// Return 404 after deletion attempt with a generic message

		}
		// Handle other errors
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error(), "type": "error"})
		return
	}

	// Return successful response
	c.JSON(http.StatusOK, gin.H{"message": "successful", "type": "success", "data": json.RawMessage(resp)})
}

// func DeleteVerificationRow(){

// }

func CreateSession(vendorString string) ([]byte, error) {
	url := "https://verification.didit.me/v1/session"
	token, err := KYCAccessToken()
	if err != nil {
		log.Println("Error Fetching Token:", err)
		return nil, err
	}

	// Parse token to check if it's valid
	var tokenData TokenResponse
	if err := json.Unmarshal(token, &tokenData); err != nil {
		return nil, fmt.Errorf("invalid token format")
	}
	if tokenData.Token == "" {
		return nil, fmt.Errorf("token not found")
	}

	// Prepare the body for the request
	body := map[string]string{
		"features":    "OCR + NFC + FACE",
		"vendor_data": vendorString,
		"callback":    "http://127.0.0.1:5500/web/admin/index.html",
	}

	// Marshal the body to JSON
	jsonifiedBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request body: %v", err)
	}

	// Prepare request options
	requestOptions := RequestOptions{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": fmt.Sprintf("Bearer %s", tokenData.Token),
		},
	}

	// Create the HTTP request
	req, err := http.NewRequest(requestOptions.Method, url, bytes.NewBuffer(jsonifiedBody))
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %v", err)
	}

	for key, value := range requestOptions.Headers {
		req.Header.Set(key, value)
	}

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Return raw JSON response
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	return responseData, nil
}
