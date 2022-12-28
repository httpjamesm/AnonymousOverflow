package main

import (
	"anonymousoverflow/src/middleware"
	"anonymousoverflow/src/routes"

	"github.com/gin-gonic/gin"
)

func main() {

	r := gin.Default()

	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./public")

	r.Use(middleware.OptionsMiddleware())

	r.GET("/robots.txt", func(c *gin.Context) {
		c.String(200, "User-agent: *\nDisallow: /")
	})

	r.GET("/", routes.GetHome)

	r.POST("/", routes.PostHome)

	r.GET("/questions/:id/:title", routes.ViewQuestion)

	r.Run("localhost:8080")
}
