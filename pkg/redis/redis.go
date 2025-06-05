package redis

import "github.com/redis/go-redis/v9"

type Redis struct {
	*redis.Client
}

var redisInstance *Redis

func GetClient() *Redis {
	return redisInstance
}

func SetClient(client *redis.Client) {
	redisInstance = &Redis{client}
}
