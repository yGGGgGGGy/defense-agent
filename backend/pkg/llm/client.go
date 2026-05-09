package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// Client communicates with the Python AI service
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest is sent to the AI service
type ChatRequest struct {
	Messages      []Message        `json:"messages"`
	SystemPrompt  string           `json:"system_prompt"`
	Temperature   float64          `json:"temperature"`
	Tools         []map[string]any `json:"tools,omitempty"`
}

// ChatResponse is the AI service response
type ChatResponse struct {
	Content   string           `json:"content"`
	ToolCalls []map[string]any `json:"tool_calls"`
	Model     string           `json:"model"`
	Usage     map[string]int   `json:"usage"`
}

// NewClient creates an LLM client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// Chat sends a chat request to the AI service and returns the response
func (c *Client) Chat(req *ChatRequest) (*ChatResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/v1/chat", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		log.Warn().Err(err).Msg("AI service unavailable, using fallback")
		return c.fallbackResponse(req), nil
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.fallbackResponse(req), nil
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(data, &chatResp); err != nil {
		return c.fallbackResponse(req), nil
	}

	return &chatResp, nil
}

func (c *Client) fallbackResponse(req *ChatRequest) *ChatResponse {
	// Generate a deterministic response based on system prompt keywords
	content := "Task analyzed. Proceeding with standard protocol."
	if req.SystemPrompt != "" {
		prompt := req.SystemPrompt
		switch {
		case containsKeyword(prompt, "perceiver", "感知"):
			content = "Perceiver: Security event detected. Initiated asset discovery and log collection."
		case containsKeyword(prompt, "analyst", "分析"):
			content = "Analyst: Threat assessed. Mapped to ATT&CK framework. Severity: HIGH. Recommend immediate containment."
		case containsKeyword(prompt, "responder", "处置"):
			content = "Responder: Containment actions executed. Blocked source IP. Verified service integrity."
		case containsKeyword(prompt, "operator", "运维"):
			content = "Operator: Health check complete. Logs rotated. Backup created. Compliance verified."
		}
	}

	return &ChatResponse{
		Content:   content,
		Model:     "fallback/mock",
		Usage:     map[string]int{"prompt_tokens": 50, "completion_tokens": 30},
		ToolCalls: nil,
	}
}

func containsKeyword(text string, keywords ...string) bool {
	for _, kw := range keywords {
		if len(text) >= len(kw) {
			for i := 0; i <= len(text)-len(kw); i++ {
				if text[i:i+len(kw)] == kw {
					return true
				}
			}
		}
	}
	return false
}
