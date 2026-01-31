package helpers

import (
	"context"
	"math/rand"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// Source - https://stackoverflow.com/a/31832326
// Posted by icza, modified by community. See post 'Timeline' for change history
// Retrieved 2026-01-31, License - CC BY-SA 4.0

func genShortUrl() string {
	// RandStringBytesMaskImprSrcSB
	const n = 10
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const (
		letterIdxBits = 6                    // 6 bits to represent a letter index
		letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
		letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
	)
	var src = rand.NewSource(time.Now().UnixNano())

	sb := strings.Builder{}
	sb.Grow(n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			sb.WriteByte(letterBytes[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return sb.String()
}

func GetShortURL(c context.Context, client *redis.Client, url string) (string, error) {

	shortUrl := genShortUrl()
	storedInDb, err := StoreShortUrl(c, client, shortUrl, url)
	return storedInDb, err
}
