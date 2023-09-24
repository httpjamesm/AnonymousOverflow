package middleware

import (
	"github.com/gin-gonic/gin"
)

func OptionsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("disable_images", false)
		c.Set("theme", "dark")

		imagesCookie, err := c.Cookie("disable_images")
		if err == nil {
			if imagesCookie == "true" {
				c.Set("disable_images", true)
			}
		}

		themeCookie, err := c.Cookie("theme")
		if err == nil {
			if themeCookie == "light" {
				c.Set("theme", "light")
			}
		}

		c.Next()
	}
}
