package routes

import (
	"net/http"
	"time"

	"github.com/ebinesh25/intolink-golang/helpers"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type request struct {
	URL string `json:"url"`
}

type response struct {
	URL            string        `json:"url"`
	XRateRemaining int           `json:"rate_limit"`
	XRateLimitRest time.Duration `json:"rate_limit_reset"`
}

func ShortenURL(url string, client *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		urled, err := helpers.GetShortURL(c, client, url)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"url": urled,
		})
	}
}
