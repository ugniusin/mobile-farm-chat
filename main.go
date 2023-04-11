package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/ugniusin/mobile-farm-chat/handlers/connection_init"
	"github.com/ugniusin/mobile-farm-chat/handlers/subscribe"
	"golang.org/x/net/websocket"
)

func main() {
	http.Handle("/graphql", websocket.Handler(subscriptionsHandler))
	log.Println("WebSocket server started at :8080/graphql")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

var messageTypes = map[string]interface{}{
	"connection_init": connection_init.Handle,
	"subscribe":       subscribe.Handle,
}

func subscriptionsHandler(ws *websocket.Conn) {
	defer ws.Close()

	log.Println("Start Subscriptions Handler")

	// Loop continuously and listen for incoming messages from the client
	for {
		var message string
		err := websocket.Message.Receive(ws, &message)

		log.Println("WebSocket message received:", message)

		if err != nil {
			if err.Error() == "EOF" {
				// Ignore the error if the client disconnected
				return
			}

			log.Println("Failed to receive message:", err)
			return
		}

		var messageMap map[string]interface{}
		messageErr := json.Unmarshal([]byte(message), &messageMap)
		if messageErr != nil {
			// Handle error
			log.Println("Message error:", messageErr)
		}

		ctx, cancel := context.WithCancel(context.Background())

		// Make sure to cancel the context when the Subscribe method returns
		defer cancel()

		messageTypes[messageMap["type"].(string)].(func(context.Context, *websocket.Conn, string))(ctx, ws, message)
	}
}
