package graphql

import (
	"log"

	"github.com/graphql-go/graphql"
	"github.com/ugniusin/mobile-farm-chat/utils/redis"
)

var memoizedSchema *graphql.Schema

func Schema() graphql.Schema {
	if memoizedSchema != nil {
		return *memoizedSchema
	}

	subscriptionMessageType := graphql.NewObject(graphql.ObjectConfig{
		Name: "SubscriptionMessage",
		Fields: graphql.Fields{
			"payload": &graphql.Field{
				Type: graphql.String,
			},
		},
	})

	// Define a GraphQL schema
	subscriptionFields := graphql.Fields{
		"events": &graphql.Field{
			Type: subscriptionMessageType,
			Args: graphql.FieldConfigArgument{
				"channel": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return struct {
					Payload interface{}
				}{
					Payload: p.Source,
				}, nil
			},
			Subscribe: func(p graphql.ResolveParams) (interface{}, error) {
				pubsub := redis.Channel(p.Args["channel"].(string))

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
	schema, _ := graphql.NewSchema(schemaConfig)

	memoizedSchema = &schema

	return schema
}
