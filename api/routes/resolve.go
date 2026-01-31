package routes

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/redis/go-redis/v9"
    "github.com/ebinesh25/intolink-golang/helpers"
)

func ResolveURL(client *redis.Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        short := c.Param("url")

        val, err := helpers.GetOriginalURL(c, client, short)
        if err != nil {
            c.JSON(http.StatusNotFound, gin.H{
                "message": "Cannot Find the URL",
            })
            return 
        }

        c.Redirect(http.StatusMovedPermanently, val)
    }
}