package connection_init

import (
	"context"
	"encoding/json"
	"log"

	"golang.org/x/net/websocket"
)

func Handle(ctx context.Context, ws *websocket.Conn, message string) {

	var m ConnectionInitMessage
	err := json.Unmarshal([]byte(message), &m)
	if err != nil {
		// Handle error
		log.Println("Message handling error: ", err)
	}

	result := ConnectionAckMessage{
		Type:    "connection_ack",
		Payload: nil,
	}

	err = websocket.JSON.Send(ws, result)

	r, _ := json.Marshal(result)

	log.Println("WebSocket message sent: ", string(r))

	if err != nil {
		log.Printf("failed to send subscription result, error: %v", err)
		// break
	}
}
