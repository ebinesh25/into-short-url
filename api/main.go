package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/ebinesh25/intolink-golang/database"
	"github.com/ebinesh25/intolink-golang/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func setupRoutes(r *gin.Engine, client *redis.Client) {

	r.GET("/*url", func(c *gin.Context) {
		// Param will return the string WITH the leading slash (e.g., "/https://google.com")
		// We trim the leading slash to get the clean URL
		url := strings.TrimPrefix(c.Param("url"), "/")

		// If there's a raw query string, append it to reconstruct URLs with query params
		// e.g., "youtube.com/watch?v=xyz" where "?v=xyz" gets parsed as request query
		if rawQuery := c.Request.URL.RawQuery; rawQuery != "" {
			url = url + "?" + rawQuery
		}

		if url == "ping" {
			c.JSON(200, gin.H{
				"message": "pong",
			})
		}

		if strings.HasPrefix(url, "http") {
			routes.ShortenURL(c, url, client)
			return
		} else {
			routes.ResolveURL(c, client, url)
			return
		}
	})

}

func setupGin(client *redis.Client) *gin.Engine {

	r := gin.Default()
	setupRoutes(r, client)
	return r
}

func setupRedis() *redis.Client {
	redisClient := database.InitRedis()
	database.TestRedis(redisClient)

	return redisClient
}

func main() {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	redis := setupRedis()
	r := setupGin(redis)

	fmt.Printf("Server starting on port %s...\n", port)

	r.Run(":" + port)

}
