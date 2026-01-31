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

func StoreShortUrl(
    ctx context.Context,
    client *redis.Client,
    short string,
    full string,
) (string, error) {
    
    _, err := client.HSet(ctx, "shortenUrls", short, full).Result()
    if err != nil {
        return "", err
    }
    
    return short, nil // Return the short string manually
}