package thirdparty

import (
	"encoding/json"
	"io"
	"log"
	"math"
	"net/http"
	"os"

	"github.com/everapihq/freecurrencyapi-go"
	"github.com/joho/godotenv"
)

// Response struct to match JSON response structure
type FreeCurResponse struct {
	Data map[string]float64 `json:"data"`
}
type ForexResponse struct {
	Success bool
	Base    string
	Time    int
	Rates   map[string]float64
}

func ForexApi() ForexResponse {

	resp, err := http.Get("https://api.forexrateapi.com/v1/latest?api_key=1895bce0d6c89851b5292e63f3ee857f&base=BTN&currencies=AUD,CAD,DKK,EUR,HKD,JPY,NOK,GBP,SGD,SEK,CHF,USD")
	if err != nil {
		log.Fatal("HTTP request failed:", err)
	}
	defer resp.Body.Close()

	// Read entire response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Failed to read response body:", err)
	}

	log.Print("Raw Json for Forexrateapi")

	var data ForexResponse
	if err := json.Unmarshal(body, &data); err != nil {
		log.Fatal("Failed to parse JSON:", err)
	}

	for ind, val := range data.Rates {
		data.Rates[ind] = 1 / val
	}
	log.Print("Converted Base and Output Generated")
	return data
}

// Load API key from .env and fetch latest forex rates
func getForex() []byte {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env: %v", err)
	}

	apiKey := os.Getenv("FOREX_API")
	if apiKey == "" {
		log.Fatal("Could not fetch API key")
	}

	// Initialize API client
	freecurrencyapi.Init(apiKey)

	// Define query parameters
	latest := map[string]string{
		"base_currency": "INR",
		"currencies":    "INR,AUD,CAD,DKK,EUR,HKD,JPY,NOK,GBP,SGD,SEK,CHF,USD",
	}

	// Fetch forex rates
	response := freecurrencyapi.Latest(latest)
	if len(response) == 0 {
		log.Println("Failed to retrieve forex data")
		return nil
	}

	log.Println("Successfully retrieved forex data")
	return response
}

// Extract only the required currencies from the full response
func getRequiredCurrency(r FreeCurResponse) FreeCurResponse {
	// Define required currencies
	currencies := []string{
		"INR", "AUD", "CAD", "DKK", "EUR", "HKD", "JPY", "NOK", "GBP", "SGD", "SEK", "CHF", "USD",
	}

	// Initialize Response struct
	resp := FreeCurResponse{
		Data: make(map[string]float64),
	}

	// Filter required currencies
	for _, cur := range currencies {
		if value, exists := r.Data[cur]; exists {
			resp.Data[cur] = math.Round((r.Data["INR"]/value)*100) / 100
		}
	}

	log.Println("Successfully filtered required currencies")
	return resp
}

// Convert API response from bytes to Response struct
func convertResponse(b []byte) FreeCurResponse {
	var vals FreeCurResponse
	err := json.Unmarshal(b, &vals)
	if err != nil {
		log.Fatalf("Could not parse response: %v", err)
	}
	log.Println("Successfully converted API response")
	return vals
}

// function to fetch and convert BTN exchange rates
func FreeCurApi() FreeCurResponse {
	rawData := getForex()
	if rawData == nil {
		log.Println("Forex API response is empty")
		return FreeCurResponse{}
	}

	fullResponse := convertResponse(rawData)

	filteredResponse := getRequiredCurrency(fullResponse)

	log.Print("Filtered Forex Data for FreeCurrencyApi")

	return filteredResponse
}
