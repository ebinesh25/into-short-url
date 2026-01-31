package routes

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/redis/go-redis/v9"
)

func ResolveURL(client *redis.Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        short := c.Param("url")

        val, err := client.Get(c, short).Result()
        if err != nil {
            c.JSON(http.StatusNotFound, gin.H{
                "message": "Cannot Find the URL",
            })
            return 
        }

        c.JSON(http.StatusOK, gin.H{
            "message": "Resolved URL",
            "short":   val,
        })
    }
}