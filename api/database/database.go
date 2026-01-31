package database

import (
	"context"
	"fmt"
	"os"
	
	"github.com/redis/go-redis/v9"
	
)


func InitRedis() *redis.Client {
	redis_url := os.Getenv("REDIS_STRING")
	if redis_url == "" {
			panic("REDIS_STRING not set")
	}
	
	opt, err := redis.ParseURL(redis_url)
	if err != nil {
		panic(err)
	}
	
	client := redis.NewClient(opt)
	
	if err := client.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}
	
	fmt.Println("Redis Connected")
	return client
}

func TestRedis(client *redis.Client) {
	ctx := context.Background()
	
	err := client.Set(ctx, "foo", "World", 0).Err()
	if err != nil {
		panic(err)
	}
	
	val, err := client.Get(ctx, "foo").Result()
	if err != nil {
		panic(err)
	}
	
	fmt.Println("foo", val)
}


