package main

import (
	"anonymousoverflow/src/middleware"
	"anonymousoverflow/src/routes"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {

	r := gin.Default()

	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./public")

	r.Use(gin.Recovery())
	r.Use(middleware.OptionsMiddleware())
	r.Use(middleware.Ratelimit())

	r.GET("/robots.txt", func(c *gin.Context) {
		c.String(200, "User-agent: *\nDisallow: /")
	})

	r.GET("/options/:name", routes.ChangeOptions)

	r.GET("/", routes.GetHome)

	r.POST("/", routes.PostHome)

	r.GET("/questions/:id/:title", routes.ViewQuestion)

	host := os.Getenv("HOST")
	if host == "" {
		host = "0.0.0.0"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(fmt.Sprintf("%s:%s", host, port))
}
