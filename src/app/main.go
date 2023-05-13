package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

type SerpResult struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

func ScrapeSerp(query string, nPages int) []SerpResult {
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.61 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/93.0.4577.63 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.131 Safari/537.36",
	}

	serpResults := make([]SerpResult, 0)
	userAgentIndex := 0

	for i := 1; i <= nPages; i++ {
		// Construct the URL for Google search
		url := fmt.Sprintf("http://www.google.com/search?q=%s&num=100&start=%d&gl=us", query, (i-1)*10)
		client := &http.Client{}
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Fatal(err)
		}

		// Set a user agent header for each request
		userAgent := userAgents[userAgentIndex]
		req.Header.Set("User-Agent", userAgent)
		userAgentIndex = (userAgentIndex + 1) % len(userAgents)

		res, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer res.Body.Close()

		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			log.Fatal(err)
		}

		// Extract search results from the HTML document
		doc.Find("div.yuRUbf").Each(func(i int, s *goquery.Selection) {
			result := SerpResult{}
			result.URL, _ = s.Find("a").Attr("href")
			result.Title = s.Find("h3").First().Text()

			serpResults = append(serpResults, result)
		})
	}

	return serpResults
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	nPagesStr := r.URL.Query().Get("n_pages")
	nPages, err := strconv.Atoi(nPagesStr)
	if err != nil {
		http.Error(w, "Invalid n_pages parameter", http.StatusBadRequest)
		return
	}

	// Scrape search results based on the provide query
	results := ScrapeSerp(query, nPages)

	// Encode results as JSON
	jsonData, err := json.Marshal(results)
	if err != nil {
		http.Error(w, "Error encode JSON", http.StatusInternalServerError)
		return
	}

	// Set response headers and write JSON data to the response
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func main() {
	// Create a new ServeMux
	mux := http.NewServeMux()

	// Register the searchHandler function for the "/search" endpoint
	mux.HandleFunc("/search", searchHandler)

	log.Printf("Starting server...")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
