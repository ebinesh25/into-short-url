package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/ebinesh25/intolink-golang/internal/database"
	"github.com/ebinesh25/intolink-golang/internal/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

type shortenRequest struct {
	URL string `json:"url" binding:"required"`
}

func setupRoutes(router *gin.Engine, client *redis.Client) {
	router.POST("/api/shorten", func(c *gin.Context) {
		var request shortenRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
			return
		}
		routes.ShortenURL(c, request.URL, client)
	})

	router.GET("/*url", func(c *gin.Context) {
		url := strings.TrimPrefix(c.Param("url"), "/")
		if url == "ping" {
			c.JSON(http.StatusOK, gin.H{"message": "pong"})
			return
		}
		url = strings.Replace(url, "http:/", "http://", 1)
		url = strings.Replace(url, "https:/", "https://", 1)

		if strings.HasPrefix(url, "http") {
			c.HTML(http.StatusOK, "redirect.html", nil)
			return
		}
		routes.ResolveURL(c, client, url)
	})
}

func setupGin(client *redis.Client) *gin.Engine {
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://send-hugss.netlify.app"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))
	router.LoadHTMLGlob("web/templates/*")
	setupRoutes(router, client)
	return router
}

func main() {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	client := database.InitRedis()
	defer client.Close()

	fmt.Printf("Server starting on port %s...\n", port)
	if err := setupGin(client).Run(":" + port); err != nil {
		panic(err)
	}
}
