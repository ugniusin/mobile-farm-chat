package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func Channel(channelName string) *redis.PubSub {
	return Client().Subscribe(context.Background(), channelName)
}
