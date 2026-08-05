package routes

import (
	"github.com/ebinesh25/intolink-golang/internal/helpers"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"net/http"
)

func ResolveURL(c *gin.Context, client *redis.Client, shortUrl string) {

	val, err := helpers.GetOriginalURL(c, client, shortUrl)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Cannot Find the URL",
		})
		return
	}

	// Increment the resolve counter
	helpers.IncrementResolveCounter(c, client, shortUrl)

	c.Redirect(http.StatusMovedPermanently, val)
}
