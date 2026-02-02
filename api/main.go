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

type ShortenRequest struct {
	URL string `json:"url"`
}

func setupRoutes(r *gin.Engine, client *redis.Client) {

	r.GET("/*url", func(c *gin.Context) {
		// Param will return the string WITH the leading slash (e.g., "/https://google.com")
		// We trim the leading slash to get the clean URL

		url := strings.TrimPrefix(c.Param("url"), "/")

		if url == "ping" {
			c.JSON(200, gin.H{
				"message": "pong",
			})
			return
		}

		if strings.HasPrefix(url, "http") {
			// render the html here
			c.HTML(200, "redirect.html", nil)
			// routes.ShortenURL(c, url, client)
			return
		} else {
			routes.ResolveURL(c, client, url)
			return
		}
	})

	r.POST("/api/shorten", func(c *gin.Context) {

		var req ShortenRequest

		if err := c.BindJSON(&req); err != nil {
			c.JSON(400, gin.H{
				"error": "invalid JSON",
			})
			return
		}

		url := req.URL
		routes.ShortenURL(c, url, client)
	})

}

func setupGin(client *redis.Client) *gin.Engine {

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
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
