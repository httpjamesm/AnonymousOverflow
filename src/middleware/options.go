package middleware

import (
	"github.com/gin-gonic/gin"
)

func OptionsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("disable_images", false)

		imagesCookie, err := c.Cookie("disable_images")
		if err == nil {
			if imagesCookie == "true" {
				c.Set("disable_images", true)
			}
		}

		c.Next()
	}
}
