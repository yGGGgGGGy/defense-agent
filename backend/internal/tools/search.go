package tools

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func searchTool() *Tool {
	return &Tool{
		Name:        "web_search",
		Description: "Search the web for security-related information (CVE, threat intel, documentation)",
		Parameters: []Param{
			{Name: "query", Type: "string", Description: "Search query", Required: true},
			{Name: "limit", Type: "integer", Description: "Max results (default 5)", Required: false},
		},
		Sandbox:   false,
		RiskLevel: "low",
		Handler: func(args map[string]string) (string, error) {
			query := args["query"]
			if query == "" {
				return "", fmt.Errorf("query is required")
			}
			return duckDuckGoSearch(query)
		},
	}
}

func duckDuckGoSearch(query string) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	apiURL := fmt.Sprintf("https://api.duckduckgo.com/?q=%s&format=json&no_html=1&skip_disambig=1",
		url.QueryEscape(query))

	resp, err := client.Get(apiURL)
	if err != nil {
		return fmt.Sprintf("Search result for '%s': connection error, try alternative search", query), nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	text := string(body)

	// Extract relevant info from DuckDuckGo response
	var results []string
	if abs := extractJSONField(text, "AbstractText"); abs != "" {
		results = append(results, abs)
	}
	if ans := extractJSONField(text, "Answer"); ans != "" {
		results = append(results, ans)
	}
	for _, topic := range extractJSONArray(text, "RelatedTopics") {
		if text := extractJSONField(topic, "Text"); text != "" {
			results = append(results, text)
			if len(results) >= 5 {
				break
			}
		}
	}

	if len(results) == 0 {
		return fmt.Sprintf("No results found for '%s'. Try a different query.", query), nil
	}

	return fmt.Sprintf("Search results for '%s':\n%s", query, strings.Join(results, "\n- ")), nil
}

func extractJSONField(json, field string) string {
	key := `"` + field + `":"`
	start := strings.Index(json, key)
	if start == -1 {
		key = `"` + field + `": "`
		start = strings.Index(json, key)
	}
	if start == -1 {
		return ""
	}
	start += len(key)
	if json[start] == '"' {
		start++
	}
	end := strings.IndexByte(json[start:], '"')
	if end == -1 {
		end = strings.IndexByte(json[start:], '}')
		if end == -1 {
			return json[start:]
		}
	}
	return strings.ReplaceAll(json[start:start+end], `\u`, ``)
}

func extractJSONArray(json, field string) []string {
	key := `"` + field + `":[`
	start := strings.Index(json, key)
	if start == -1 {
		return nil
	}
	start += len(key)
	depth := 1
	end := start
	for end < len(json) && depth > 0 {
		if json[end] == '[' {
			depth++
		} else if json[end] == ']' {
			depth--
		}
		end++
	}
	raw := json[start : end-1]

	// Split by "},{"
	items := strings.Split(raw, "},{")
	var result []string
	for _, item := range items {
		result = append(result, "{"+strings.Trim(item, "{}")+"}")
	}
	return result
}
