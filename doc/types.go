package bs

import (
	"fmt"
	"time"
)

// possible structures to help the middle with the
// backend and the user

// InitialContext represents the basic context structure created from a user prompt.
type InitialContext struct {
	UserPrompt string                 `json:"userPrompt"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// EnhancedContext represents a more detailed structure for maintaining state.
type EnhancedContext struct {
	UserID          string                   `json:"userID"`
	SessionID       string                   `json:"sessionID"`
	LastUpdated     string                   `json:"lastUpdated"`
	SessionStart    string                   `json:"sessionStart"`
	History         []map[string]interface{} `json:"history"`
	Entities        []string                 `json:"entities"`
	Concepts        []string                 `json:"concepts"`
	CurrentState    string                   `json:"currentState"`
	Expectations    string                   `json:"expectations"`
	OpenAIParams    map[string]interface{}   `json:"openAIParams"`
	References      []string                 `json:"references"`
	UserPreferences map[string]interface{}   `json:"userPreferences"`
	Metadata        map[string]interface{}   `json:"metadata"`
}

// NewInitialContext creates a new InitialContext from a user prompt.
func NewInitialContext(userPrompt string) *InitialContext {
	return &InitialContext{
		UserPrompt: userPrompt,
		Metadata: map[string]interface{}{
			"lastUpdated": time.Now().Format(time.RFC3339),
			"changeHistory": []string{
				fmt.Sprintf("Initial prompt received: %s", time.Now().Format(time.RFC3339)),
			},
		},
	}
}

// NewEnhancedContext creates a new EnhancedContext with default values and the provided user ID and session ID.
func NewEnhancedContext(userID, sessionID string) *EnhancedContext {
	return &EnhancedContext{
		UserID:       userID,
		SessionID:    sessionID,
		LastUpdated:  time.Now().Format(time.RFC3339),
		SessionStart: time.Now().Format(time.RFC3339),
		History:      []map[string]interface{}{},
		Entities:     []string{},
		Concepts:     []string{},
		CurrentState: "awaiting_response",
		Expectations: "none",
		OpenAIParams: map[string]interface{}{
			"temperature": 0.7,
			"maxTokens":   150,
		},
		References: []string{},
		UserPreferences: map[string]interface{}{
			"language":       "English",
			"formalityLevel": "casual",
		},
		Metadata: map[string]interface{}{
			"changeHistory": []string{
				fmt.Sprintf("Session started: %s", time.Now().Format(time.RFC3339)),
			},
		},
	}
}
