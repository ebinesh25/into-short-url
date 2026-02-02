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

    // Store reverse mapping for duplicate checking
    _, err = client.HSet(ctx, "originalUrls", full, short).Result()
    if err != nil {
        return "", err
    }

    return short, nil // Return the short string manually
}

// Check if original URL already exists and return its short URL
func GetShortURLByOriginal(
    ctx context.Context,
    client *redis.Client,
    originalURL string,
) (string, error) {

    return client.HGet(ctx, "originalUrls", originalURL).Result()
}

func IncrementResolveCounter(ctx context.Context, client *redis.Client, short string) error {
    _, err := client.HIncrBy(ctx, "resolveCounter", short, 1).Result()
    return err
}