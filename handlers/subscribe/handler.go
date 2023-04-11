package subscribe

import (
	"context"
	"encoding/json"
	"log"

	graphqlgo "github.com/graphql-go/graphql"
	"github.com/ugniusin/mobile-farm-chat/utils/graphql"
	"golang.org/x/net/websocket"
)

func Handle(ctx context.Context, ws *websocket.Conn, message string) {
	var m SubscribeMessage
	err := json.Unmarshal([]byte(message), &m)
	if err != nil {
		// Handle error
		log.Println("Message handling error: ", err)
	}

	requestString := m.Payload.Query
	operationId := m.ID

	log.Println("Operation ID: ", operationId)

	// Parse the incoming message as a GraphQL query
	params := graphqlgo.Params{Schema: graphql.Schema(), RequestString: requestString, Context: ctx}

	subscription := graphqlgo.Subscribe(params)

	go func() {
		// Loop over the subscription and write results to the WebSocket
		for result := range subscription {
			//err = websocket.JSON.Send(ws, result)

			r, _ := json.Marshal(result)

			response := NextMessage{
				ID:      operationId,
				Type:    "next",
				Payload: *result,
			}

			err = websocket.JSON.Send(ws, response)

			log.Println("WebSocket message sent: ", string(r))

			if err != nil {
				log.Printf("failed to send subscription result, error: %v", err)
				break
			}
		}
	}()
}
