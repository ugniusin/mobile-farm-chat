package main

import (
	"context"
  "fmt"
	"log"
	"net/http"
	"sync"

	"github.com/graphql-go/graphql"
	"golang.org/x/net/websocket"
)

var schema graphql.Schema

func init() {
	// Define a GraphQL schema
	fields := graphql.Fields{
		"counter": &graphql.Field{
			Type: graphql.String,
			Args: graphql.FieldConfigArgument{
				"count": &graphql.ArgumentConfig{
					Type: graphql.Int,
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				// Extract the count argument from the resolver params
				count, ok := p.Args["count"].(int)
				if !ok {
					count = 1
				}

				// Generate a greeting message with the specified count
				greeting := ""
				for i := 0; i < count; i++ {
					greeting += "Hello, world!\n"
				}

				return greeting, nil
			},
			Subscribe: func(p graphql.ResolveParams) (interface{}, error) {
				// Extract the count argument from the resolver params
				count, ok := p.Args["count"].(int)
				if !ok {
					count = 1
				}

				// Return a channel that generates the greeting message with the specified count
				c := make(chan interface{})
				go func() {
					for i := 0; i < count; i++ {
						c <- fmt.Sprintf("Hello, world! [%d]", i+1)
					}
					close(c)
				}()
				return c, nil
			},
		},
	}
	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: fields}
	rootMutation := graphql.ObjectConfig{Name: "RootMutation", Fields: fields}
	rootSubscription := graphql.ObjectConfig{Name: "RootSubscription", Fields: fields}

	schemaConfig := graphql.SchemaConfig{
		Query: graphql.NewObject(rootQuery),
		Mutation: graphql.NewObject(rootMutation),
		Subscription: graphql.NewObject(rootSubscription),
	}
	schema, _ = graphql.NewSchema(schemaConfig)
}

func main() {
	// Initialize a mutex to synchronize access to the subscription manager
	// var mu sync.Mutex

	// Initialize a map to store the active subscriptions
	// subscriptions := make(map[*websocket.Conn]bool)

	http.Handle("/graphql", websocket.Handler(subscriptionsHandler))
	log.Println("WebSocket server started at :8080/graphql")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func graphqlHandler(ws *websocket.Conn) {
	defer ws.Close()

	var message string
	err := websocket.Message.Receive(ws, &message)
	if err != nil {
		log.Println("Failed to receive message:", err)
		return
	}

	// Execute the GraphQL query
	params := graphql.Params{Schema: schema, RequestString: message}
	result := graphql.Do(params)

	// Send the result back to the client
	err = websocket.JSON.Send(ws, result)
	if err != nil {
		log.Println("Failed to send result:", err)
	}
}


func subscriptionsHandler(ws *websocket.Conn) {
	// Add the WebSocket connection to the subscription manager
	//mu.Lock()
	//subscriptions[ws] = true
	//mu.Unlock()

	// Set up a GraphQL subscription manager
	//subMgr := graphql.NewSubscriptionManager(&schema)

	// Close the WebSocket connection and remove it from the subscription manager when the client disconnects
	//defer func() {
	//	ws.Close()
	//	mu.Lock()
	//	delete(subscriptions, ws)
	//	mu.Unlock()
	//}()

	ctx := context.WithValue(context.Background(), "counter", 0)

	defer ws.Close()
	
	// GraphQL WebSocket message format
	type graphqlWSMessage struct {
		Type    string                 `json:"type"`
		ID      string                 `json:"id,omitempty"`
		Payload map[string]interface{} `json:"payload"`
	}

	// Loop continuously and listen for incoming messages from the client
	for {
		log.Println("LOOP")

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

		// Parse the incoming message as a GraphQL query
		// params := graphql.Params{Schema: schema, RequestString: message}

		params := graphql.Params{Schema: schema, RequestString: message, Context: ctx}

		// Parse the incoming message as a GraphQL subscription query
		//doc, err := parser.Parse(parser.ParseParams{Source: source.NewSource(&source.Source{Body: message})})
		//if err != nil {
		//	log.Println("Failed to parse subscription:", err)
		//	continue
		//}

		// Validate the subscription query against the schema
		//validationResult := graphql.ValidateDocument(&schema, doc, nil)
		//if !validationResult.IsValid {
		//	errMsgs := gqlerrors.FormatErrors(validationResult.Errors)
		//	log.Println("Failed to validate subscription:", errMsgs)
		//	continue
		//}
		
		// If the operation is a subscription, execute it in a separate goroutine and send the results back to the client
		if params.OperationName == "subscription" {
			go func() {
				// result := graphql.Execute(params)
				//result := graphql.Do(params)

				subscriptionChannel := graphql.Subscribe(params)


				// defer subscription.Unsubscribe()

				// Create a wait group to ensure that the subscription channel is closed before exiting the function
				wg := sync.WaitGroup{}
				wg.Add(1)
				defer wg.Wait()

				// Loop over the subscription and write results to the WebSocket
				for result := range subscriptionChannel {
					err = websocket.JSON.Send(ws, result)
					if err != nil {
						log.Printf("failed to send subscription result, error: %v", err)
						break
					}
				}

				// Mark the wait group as done
				wg.Done()
			}()
		} else {
			// If the operation is not a subscription, execute it synchronously and send the results back to the client
		
			// var message string
			//err := websocket.Message.Receive(ws, &message)
			//if err != nil {

			//	log.Println("ELSE ERROR:", err)

			//	log.Println("Failed to receive message:", err)
			//	return
			//}

			// Execute the GraphQL query
			//params := graphql.Params{Schema: schema, RequestString: message}
			result := graphql.Do(params)

			// Send the result back to the client
			err = websocket.JSON.Send(ws, result)
			if err != nil {
				log.Println("Failed to send result:", err)
			}
		}
	}
}
