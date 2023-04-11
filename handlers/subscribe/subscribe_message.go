package subscribe

type SubscribeMessage struct {
	ID      string  `json:"id"`
	Type    string  `json:"type"`
	Payload Payload `json:"payload"`
}

type Payload struct {
	OperationName string                 `json:"operationName,omitempty"`
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	Extensions    map[string]interface{} `json:"extensions,omitempty"`
}
