package subscribe

import "github.com/graphql-go/graphql"

type NextMessage struct {
	ID      string         `json:"id"`
	Type    string         `json:"type"`
	Payload graphql.Result `json:"payload"`
	// Payload ExecutionResult `json:"payload"`
}

type ExecutionResult struct {
	Data       string `json:"data,omitempty"`
	Extensions string `json:"Extensions,omitempty"`
}
