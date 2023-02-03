package main

import (
	"anonymousoverflow/env"
	"anonymousoverflow/src/middleware"
	"anonymousoverflow/src/routes"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {

	env.RunChecks()

	host := os.Getenv("HOST")
	if host == "" {
		host = "0.0.0.0"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if os.Getenv("DEV") != "true" {
		gin.SetMode(gin.ReleaseMode)
		fmt.Printf("Running in production mode. Listening on %s:%s.", host, port)
	}

	r := gin.Default()

	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./public")

	r.Use(gin.Recovery())
	r.Use(middleware.XssPreventionHeaders())
	r.Use(middleware.NoCacheMiddleware())
	r.Use(middleware.OptionsMiddleware())
	r.Use(middleware.Ratelimit())

	r.GET("/robots.txt", func(c *gin.Context) {
		c.String(200, "User-agent: *\nDisallow: /")
	})

	r.GET("/options/:name", routes.ChangeOptions)

	r.GET("/", routes.GetHome)

	r.POST("/", routes.PostHome)

	r.GET("/questions/:id/:title", routes.ViewQuestion)

	r.GET("/proxy", routes.GetImage)

	r.Run(fmt.Sprintf("%s:%s", host, port))
}
