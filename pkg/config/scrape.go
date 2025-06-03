package config

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/playwright-community/playwright-go"
)

// Currency represents the structure of currency data.
type Currency struct {
	Name  string `json:"name"`
	Buy   string `json:"buy"`
	Sell  string `json:"sell"`
	Buy2  string `json:"buy2,omitempty"`
	Sell2 string `json:"sell2,omitempty"`
}

// scrapes currency data from multiple URLs and outputs it as JSON.
func Scrape() map[string][]Currency {
	urls := []string{
		"https://www.rma.org.bt/exchangeRates/",
		"https://www.bob.bt/",
		"https://bnb.bt/forex/",
	}

	pw, browser := initPlaywright()
	defer pw.Stop()
	defer browser.Close()

	var (
		wg          sync.WaitGroup
		mu          sync.Mutex
		currencyMap = make(map[string][]Currency) // Map to store currencies by page title
	)

	// Scrape each URL concurrently
	for _, url := range urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()

			currencies, err := scrapePage(browser, u)
			if err != nil {
				log.Printf("Error scraping %s: %v", u, err)
				return
			}

			mu.Lock()
			currencyMap[url] = currencies
			mu.Unlock()
		}(url)
	}

	wg.Wait() // Wait for all goroutines to finish

	// Output the results as JSON
	return currencyMap
}

// initializes Playwright and launches a browser.
func initPlaywright() (*playwright.Playwright, playwright.Browser) {
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("Failed to start Playwright: %v", err)
	}

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		log.Fatalf("Failed to launch browser: %v", err)
	}

	return pw, browser
}

// scrapes currency data from a single URL and returns the currencies and page title.
func scrapePage(browser playwright.Browser, url string) ([]Currency, error) {
	page, err := browser.NewPage()
	if err != nil {
		return nil, fmt.Errorf("failed to create page: %w", err)
	}
	defer page.Close()

	if _, err := page.Goto(url); err != nil {
		return nil, fmt.Errorf("failed to navigate to %s: %w", url, err)
	}

	if _, err := page.WaitForSelector("table"); err != nil {
		return nil, fmt.Errorf("failed to wait for table on %s: %w", url, err)
	}

	log.Printf("Scraping page: %s", url)

	table := page.Locator("table")
	rows, err := table.Locator("tbody tr").All()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows: %w", err)
	}

	var currencies []Currency

	for _, row := range rows {
		cells, err := row.Locator("td").All()
		if err != nil {
			log.Printf("Failed to get cells for row: %v", err)
			continue
		}

		if len(cells) < 3 {
			continue //
		}

		currency, err := extractCurrencyData(cells)
		if err != nil {
			log.Printf("Failed to extract currency data: %v", err)
			continue
		}

		currencies = append(currencies, currency)
	}

	return currencies, nil
}

// extracts currency data from a row of cells.
func extractCurrencyData(cells []playwright.Locator) (Currency, error) {
	name, err := cells[0].TextContent()
	if err != nil {
		return Currency{}, fmt.Errorf("failed to get name: %w", err)
	}
	name = cleanString(name)

	buy, err := cells[1].TextContent()
	if err != nil {
		return Currency{}, fmt.Errorf("failed to get buy value: %w", err)
	}
	buy = cleanString(buy)

	sell, err := cells[2].TextContent()
	if err != nil {
		return Currency{}, fmt.Errorf("failed to get sell value: %w", err)
	}
	sell = cleanString(sell)

	currency := Currency{
		Name: name,
		Buy:  buy,
		Sell: sell,
	}

	if len(cells) >= 5 {
		buy2, err := cells[3].TextContent()
		if err != nil {
			return Currency{}, fmt.Errorf("failed to get buy2 value: %w", err)
		}
		buy2 = cleanString(buy2)

		sell2, err := cells[4].TextContent()
		if err != nil {
			return Currency{}, fmt.Errorf("failed to get sell2 value: %w", err)
		}
		sell2 = cleanString(sell2)

		currency.Buy2 = buy2
		currency.Sell2 = sell2
	}

	return currency, nil
}

// cleans up a string by removing unwanted characters.
func cleanString(s string) string {
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.TrimSpace(s)
	s = strings.Replace(s, "                                                                            ", ":-", 1)

	return s
}

// outputJSON converts the data to JSON and prints it.
func outputJSON(data map[string][]Currency) []byte {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}
	// fmt.Println(string(jsonData))
	log.Printf("Data Scraped: http status: %v", http.StatusFound)
	return jsonData
}
