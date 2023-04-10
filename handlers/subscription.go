package handlers

import (
	"context"
	"log"

	"github.com/graphql-go/graphql"
	"github.com/redis/go-redis/v9"
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

				log.Println("Redis subscribed")

				c := make(chan interface{})
				go func() {
					for {
						select {
						case <-p.Context.Done():
							log.Println("Redis unsubscribed")

							pubsub.Unsubscribe(p.Context, p.Args["channel"].(string))
							pubsub.Close()

							return
						case message := <-pubsub.Channel():

							log.Println("Redis message received:", message.Payload)

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

func Handle(subscribeMessage SubscribeMessage) chan *graphql.Result {
	requestString := subscribeMessage.Payload.Query

	ctx, cancel := context.WithCancel(context.Background())

	// Make sure to cancel the context when the Subscribe method returns
	defer cancel()

	// Parse the incoming message as a GraphQL query
	params := graphql.Params{Schema: schema, RequestString: requestString, Context: ctx}

	subscription := graphql.Subscribe(params)

	return subscription
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
