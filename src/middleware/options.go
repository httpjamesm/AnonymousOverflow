package middleware

import "github.com/gin-gonic/gin"

func OptionsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("disable_images", false)

		// get cookie
		cookie, err := c.Cookie("disable_images")
		if err != nil {
			c.Next()
			return
		}

		// check if disable_images is equal to "true"
		if cookie == "true" {
			c.Set("disable_images", true)
		}

		c.Next()
	}
}
