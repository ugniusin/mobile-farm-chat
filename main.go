package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"runtime"

	"github.com/graphql-go/graphql"
	"github.com/redis/go-redis/v9"
	"golang.org/x/net/websocket"
)

var schema graphql.Schema

var rdb = PubSubClient()

func init() {
	subscriptionMessageType := graphql.NewObject(graphql.ObjectConfig{
		Name: "SubscriptionMessage",
		Fields: graphql.Fields{
			"channel": &graphql.Field{
				Type: graphql.String,
			},
			"message": &graphql.Field{
				Type: graphql.String,
			},
		},
	})

	// Define a GraphQL schema
	subscriptionFields := graphql.Fields{
		"event": &graphql.Field{
			Type: subscriptionMessageType,
			Args: graphql.FieldConfigArgument{
				"channel": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return struct {
					Channel string
					Message interface{}
				}{
					Channel: p.Args["channel"].(string),
					Message: p.Source,
				}, nil
			},
			Subscribe: func(p graphql.ResolveParams) (interface{}, error) {
				pubsub := PubSubChannel(p.Args["channel"].(string))

				c := make(chan interface{})
				go func() {
					for {
						select {
						case <-p.Context.Done():
							pubsub.Unsubscribe(p.Context, p.Args["channel"].(string))
							pubsub.Close()

							return
						case message := <-pubsub.Channel():
							c <- message.Payload
						}
					}
				}()
				return c, nil
			},
		},
	}
	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: graphql.Fields{}}
	rootMutation := graphql.ObjectConfig{Name: "RootMutation", Fields: graphql.Fields{}}
	rootSubscription := graphql.ObjectConfig{Name: "RootSubscription", Fields: subscriptionFields}

	schemaConfig := graphql.SchemaConfig{
		Query:        graphql.NewObject(rootQuery),
		Mutation:     graphql.NewObject(rootMutation),
		Subscription: graphql.NewObject(rootSubscription),
	}
	schema, _ = graphql.NewSchema(schemaConfig)
}

func main() {
	http.Handle("/graphql", websocket.Handler(subscriptionsHandler))
	log.Println("WebSocket server started at :8080/graphql")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func subscriptionsHandler(ws *websocket.Conn) {
	defer ws.Close()

	defer func() {
		buf := make([]byte, 1<<20)
		runtime.Stack(buf, true)
		fmt.Println("Stack trace of all goroutines:\n", string(buf))
	}()

	// Loop continuously and listen for incoming messages from the client
	for {
		var message string
		err := websocket.Message.Receive(ws, &message)

		if err != nil {
			if err.Error() == "EOF" {
				// Ignore the error if the client disconnected
				return
			}

			log.Println("Failed to receive message:", err)
			return
		}

		ctx, cancel := context.WithCancel(context.Background())

		// Make sure to cancel the context when the Subscribe method returns
		defer cancel()

		// Parse the incoming message as a GraphQL query
		params := graphql.Params{Schema: schema, RequestString: message, Context: ctx}

		subscription := graphql.Subscribe(params)

		go func() {
			// Loop over the subscription and write results to the WebSocket
			for result := range subscription {
				err = websocket.JSON.Send(ws, result)
				if err != nil {
					log.Printf("failed to send subscription result, error: %v", err)
					break
				}
			}
		}()
	}
}

func PubSubClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "password",
		DB:       0, // use default DB
	})
}

func PubSubChannel(channelName string) *redis.PubSub {
	return rdb.Subscribe(context.Background(), channelName)
}
