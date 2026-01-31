package routes

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/redis/go-redis/v9"
    "github.com/ebinesh25/intolink-golang/helpers"
)

func ResolveURL(c *gin.Context, client *redis.Client, shortUrl string) {

    val, err := helpers.GetOriginalURL(c, client, shortUrl)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{
            "message": "Cannot Find the URL",
        })
        return 
    }

    c.Redirect(http.StatusMovedPermanently, val)
}