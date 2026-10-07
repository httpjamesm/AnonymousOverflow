package main

import (
	"anonymousoverflow/env"
	"anonymousoverflow/src/middleware"
	"anonymousoverflow/src/routes"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	healthcheck "github.com/tavsec/gin-healthcheck"
	"github.com/tavsec/gin-healthcheck/checks"
	"github.com/tavsec/gin-healthcheck/config"
)

//go:embed templates/*
var templates embed.FS

//go:embed public/*
var public embed.FS

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
		fmt.Printf("Running in production mode. Listening on %s:%s.\n", host, port)
	}

	r := gin.Default()

	// gin trusts X-Forwarded-For from any client by default, which lets clients
	// pick their own IP and bypass the rate limiter.
	trustedProxies := []string{"127.0.0.0/8", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "::1/128", "fc00::/7"}
	if v := os.Getenv("TRUSTED_PROXIES"); v != "" {
		trustedProxies = strings.Split(v, ",")
	}
	if err := r.SetTrustedProxies(trustedProxies); err != nil {
		panic(err)
	}

	templ := template.Must(template.New("").ParseFS(templates, "templates/*"))
	r.SetHTMLTemplate(templ)

	r.Use(gin.Recovery())
	r.Use(middleware.XssPreventionHeaders())
	r.Use(middleware.OptionsMiddleware())
	r.Use(middleware.Ratelimit())

	static, _ := fs.Sub(public, "public")
	r.StaticFS("/static", http.FS(static))

	r.GET("/robots.txt", func(c *gin.Context) {
		c.String(200, "User-agent: *\nDisallow: /")
	})

	r.GET("/options/:name", routes.ChangeOptions)

	r.GET("/", routes.GetHome)
	r.POST("/", routes.PostHome)

	r.GET("/a/:id", routes.RedirectShortenedOverflowURL)
	r.GET("/a/:id/:answerId", routes.RedirectShortenedOverflowURL)
	r.GET("/q/:id", routes.RedirectShortenedOverflowURL)
	r.GET("/q/:id/:answerId", routes.RedirectShortenedOverflowURL)

	exchangeRouter := r.Group("/exchange/:sub")
	{
		exchangeRouter.GET("/questions/:id/:title", routes.ViewQuestion)
		exchangeRouter.GET("/questions/:id", func(c *gin.Context) {
			// redirect user to the question with the title
			c.Redirect(302, fmt.Sprintf("/exchange/%s/questions/%s/placeholder", c.Param("sub"), c.Param("id")))
		})
		exchangeRouter.GET("/questions/:id/:title/:answerId", func(c *gin.Context) {
			// redirect user to the answer with the title
			c.Redirect(302, fmt.Sprintf("/exchange/%s/questions/%s/%s#%s", c.Param("sub"), c.Param("id"), c.Param("title"), c.Param("answerId")))
		})
		exchangeRouter.GET("/q/:id/:answerId", routes.RedirectShortenedOverflowURL)
		exchangeRouter.GET("/q/:id", routes.RedirectShortenedOverflowURL)
		exchangeRouter.GET("/a/:id/:answerId", routes.RedirectShortenedOverflowURL)
		exchangeRouter.GET("/a/:id", routes.RedirectShortenedOverflowURL)
	}

	r.GET("/questions/:id", func(c *gin.Context) {
		// redirect user to the question with the title
		c.Redirect(302, fmt.Sprintf("/questions/%s/placeholder", c.Param("id")))
	})
	r.GET("/questions/:id/:title", routes.ViewQuestion)
	r.GET("/questions/:id/:title/:answerId", func(c *gin.Context) {
		// redirect user to the answer with the title
		c.Redirect(302, fmt.Sprintf("/questions/%s/%s#%s", c.Param("id"), c.Param("title"), c.Param("answerId")))
	})

	r.GET("/proxy", routes.GetImage)

	r.GET("/version", routes.GetVersion)

	soPingCheck := checks.NewPingCheck("https://stackoverflow.com", "GET", 5000, nil, nil)
	sePingCheck := checks.NewPingCheck("https://stackexchange.com", "GET", 5000, nil, nil)
	healthcheck.New(r, config.DefaultConfig(), []checks.Check{soPingCheck, sePingCheck})

	r.Run(fmt.Sprintf("%s:%s", host, port))
}
