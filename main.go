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

	r.Use(gin.Recovery())
	r.Use(middleware.XssPreventionHeaders())
	r.Use(middleware.NoCacheMiddleware())
	r.Use(middleware.OptionsMiddleware())
	r.Use(middleware.Ratelimit())

	r.GET("/static/*filepath", routes.StaticContent)

	r.GET("/robots.txt", func(c *gin.Context) {
		c.String(200, "User-agent: *\nDisallow: /")
	})

	r.GET("/options/:name", routes.ChangeOptions)

	r.GET("/", routes.GetHome)
	r.POST("/", routes.PostHome)

	r.GET("/a/:id", routes.RedirectShortenedOverflowURL)
	r.GET("/q/:id", routes.RedirectShortenedOverflowURL)

	r.GET("/questions/:id", func(c *gin.Context) {
		// redirect user to the question with the title
		c.Redirect(302, fmt.Sprintf("/questions/%s/placeholder", c.Param("id")))
	})
	r.GET("/questions/:id/:title", routes.ViewQuestion)
	r.GET("/questions/:id/:title/:answerId", func(c *gin.Context) {
		// redirect user to the answer with the title
		c.Redirect(302, fmt.Sprintf("/questions/%s/%s#%s", c.Param("id"), c.Param("title"), c.Param("answerId")))
	})
	r.GET("/exchange/:sub/questions/:id/:title", routes.ViewQuestion)
	r.GET("/exchange/:sub/questions/:id", func(c *gin.Context) {
		// redirect user to the question with the title
		c.Redirect(302, fmt.Sprintf("/exchange/%s/questions/%s/placeholder", c.Param("sub"), c.Param("id")))
	})
	r.GET("/exchange/:sub/questions/:id/:title/:answerId", func(c *gin.Context) {
		// redirect user to the answer with the title
		c.Redirect(302, fmt.Sprintf("/exchange/%s/questions/%s/%s#%s", c.Param("sub"), c.Param("id"), c.Param("title"), c.Param("answerId")))
	})

	r.GET("/proxy", routes.GetImage)

	r.Run(fmt.Sprintf("%s:%s", host, port))
}
