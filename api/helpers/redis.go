package helpers

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func GetOriginalURL(
	ctx context.Context,
	client *redis.Client,
	short string,
) (string, error) {

	return client.HGet(ctx, "shortenUrls", short).Result()
}
