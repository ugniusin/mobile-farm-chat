package handlers

type SubscribeMessage struct {
	ID      string           `json:"id"`
	Type    string           `json:"type"`
	Payload SubscribePayload `json:"payload"`
}

type SubscribePayload struct {
	OperationName string                 `json:"operationName,omitempty"`
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	Extensions    map[string]interface{} `json:"extensions,omitempty"`
}
