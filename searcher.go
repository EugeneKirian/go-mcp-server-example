package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var userAgents = [...]string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Linux; Android 8.1.0; BQ-6010G Build/O11019) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Mobile Safari/537.36 YaApp_Android/10.80 YaSearchBrowser/10.80",
	"Mozilla/5.0 (Linux; Android 8.0.0; PRA-TL10) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.96 Mobile Safari/537.36",
	"Mozilla/5.0 (Linux; Android 10; Redmi Note 9 Pro Build/QKQ1.191215.002; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/92.0.4515.166 Mobile Safari/537.36",
	"Dalvik/2.1.0 (Linux; U; Android 6.0.1; Redmi 3X MIUI/V9.5.5.0.MALMIFA)",
	"Mozilla/5.0 (iPad; CPU OS 15_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) CriOS/96.0.4664.53 Mobile/15E148 Safari/604.1",
	"Dalvik/2.1.0 (Linux; U; Android 9; Redmi Note 7 Pro MIUI/V10.3.5.0.PFHINXM)",
	"Mozilla/5.0 (Linux; Android 7.0; SM-G920T1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.83 Mobile Safari/537.36",
	"Mozilla/5.0 (Linux; Android 9; STK-LX1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.106 Mobile Safari/537.36",
	"Mozilla/5.0 (iPad; CPU OS 14_8 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) EdgiOS/93.0.961.47 Version/14.0 Mobile/15E148 Safari/604.1",
	"Mozilla/5.0 (X11; U; Linux i686) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.96 Safari/537.36",
}

type Searcher struct {
	Query string `json:"query" jsonschema:"Search query"`
}

func NewSearcher(query string) *Searcher {
	return &Searcher{Query: query}
}

// extractResultsArray finds the results:[...] array inside the web:{type:"search" block
// and returns the raw JSON array string.
func extractResultsArray(body string) string {
	// Step 4.2: find the start of the web search block
	webStart := strings.Index(body, `web:{type:"search"`)
	if webStart == -1 {
		return ""
	}

	// Step 4.3: within that block, find results:[
	resultsKey := "results:["
	resultsStart := strings.Index(body[webStart:], resultsKey)
	if resultsStart == -1 {
		return ""
	}
	// Position of the opening [ of the array
	arrayStart := webStart + resultsStart + len(resultsKey) - 1

	// Step 4.4: scan forward tracking brackets to find the complete array
	depth := 0
	for i := arrayStart; i < len(body); i++ {
		switch body[i] {
		case '[':
			depth++
		case ']':
			depth--
			if depth == 0 {
				return body[arrayStart : i+1]
			}
		}
	}
	return ""
}

// parseResults parses the non-standard JS object array into Content objects.
// The array elements use unquoted keys (JS object literals), so we use a
// best-effort field extraction approach.
func parseResults(array string) []SearchResult {
	// Try standard JSON first (Brave may emit valid JSON in some responses)
	var items []SearchResult
	if err := json.Unmarshal([]byte(array), &items); err == nil {
		return items
	}

	// Fall back: extract title and url fields by scanning each object
	var results []SearchResult
	// Split on top-level object boundaries
	depth := 0
	start := -1
	for i, ch := range array {
		switch ch {
		case '{':
			if depth == 0 {
				start = i
			}
			depth++
		case '}':
			depth--
			if depth == 0 && start != -1 {
				obj := array[start : i+1]
				c := SearchResult{
					URL:   extractField(obj, "url"),
					Title: extractField(obj, "title"),
				}
				if c.URL != "" {
					results = append(results, c)
				}
				start = -1
			}
		}
	}
	return results
}

// extractField finds the value of a given key in a JS/JSON object string.
// Handles both quoted string values (e.g. title:"foo" or title:"foo").
func extractField(obj, key string) string {
	// Try key:"value" and key:"value" patterns
	for _, sep := range []string{`"` + key + `":"`, key + `:"`} {
		idx := strings.Index(obj, sep)
		if idx == -1 {
			continue
		}
		valueStart := idx + len(sep)
		// Find closing quote, respecting escaped quotes
		for i := valueStart; i < len(obj); i++ {
			if obj[i] == '"' && (i == 0 || obj[i-1] != '\\') {
				return obj[valueStart:i]
			}
		}
	}
	return ""
}

func (s *Searcher) Search() ([]SearchResult, error) {
	// Step 1: validate query
	if strings.TrimSpace(s.Query) == "" {
		return nil, fmt.Errorf("Query cannot be empty.")
	}

	// Step 1.3 + 2.1: join multiple args, URL-encode the query
	searchURL := "https://search.brave.com/search?q=" + url.QueryEscape(s.Query) + "&source=web"

	// Step 3: perform HTTP GET with User-Agent and timeout
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", userAgents[rand.Intn(len(userAgents))])

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Web request failed: %w", err)
	}
	defer resp.Body.Close()

	// Step 3.2: check HTTP status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Web request failed: HTTP %d", resp.StatusCode)
	}

	// Step 4.1: read the full response body as a string
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read response body: %w", err)
	}
	body := string(bodyBytes)

	// Steps 4.2–4.4: locate and extract the results JSON array
	array := extractResultsArray(body)

	// Step 5: parse the results array into Content objects
	var items []SearchResult
	if array != "" {
		items = parseResults(array)
	}

	// Step 5.3: ensure items is never null in JSON output
	if items == nil {
		items = []SearchResult{}
	}

	// Step 6: serialize
	return items, nil
}
