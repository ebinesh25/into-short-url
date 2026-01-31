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


func setupRoutes(r *gin.Engine, client *redis.Client){
	
	r.GET("/*url", func(c *gin.Context) {
		// Param will return the string WITH the leading slash (e.g., "/https://google.com")
        // We trim the leading slash to get the clean URL
        url := strings.TrimPrefix(c.Param("url"), "/")
		
        if (url == "ping") {
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

func main(){
	err := godotenv.Load()
	if err != nil {
		fmt.Println(err)
	}
	
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	redis := setupRedis()
	r := setupGin(redis)
	r.Run(":" + port)
	
	
}