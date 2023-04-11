package redis

import "github.com/redis/go-redis/v9"

var memoizedClient *redis.Client

func Client() *redis.Client {
	if memoizedClient != nil {
		return memoizedClient
	}

	client := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "password",
		DB:       0, // use default DB
	})

	memoizedClient = client
	return client
}
