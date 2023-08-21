package routes

import (
	"anonymousoverflow/config"
	"fmt"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetHome(c *gin.Context) {
	c.HTML(200, "home.html", gin.H{
		"version": config.Version,
		"theme":   c.MustGet("theme").(string),
	})
}

type urlConversionRequest struct {
	URL string `form:"url" binding:"required"`
}

var stackExchangeRegex = regexp.MustCompile(`https://(.+).stackexchange.com/questions/`)

func PostHome(c *gin.Context) {
	body := urlConversionRequest{}

	if err := c.ShouldBind(&body); err != nil {
		c.HTML(400, "home.html", gin.H{
			"errorMessage": "Invalid request body",
			"theme":        c.MustGet("theme").(string),
		})
		return
	}

	soLink := body.URL

	// validate URL
	isStackOverflow := strings.HasPrefix(soLink, "https://stackoverflow.com/questions/")
	isShortenedStackOverflow := strings.HasPrefix(soLink, "https://stackoverflow.com/a/")
	isStackExchange := stackExchangeRegex.MatchString(soLink)
	if !isStackExchange && !isStackOverflow && !isShortenedStackOverflow {
		c.HTML(400, "home.html", gin.H{
			"errorMessage": "Invalid stack overflow/exchange URL",
			"theme":        c.MustGet("theme").(string),
		})
		return
	}

	// if stack overflow, trim https://stackoverflow.com
	if isStackOverflow {
		c.Redirect(302, strings.TrimPrefix(soLink, "https://stackoverflow.com"))
		return
	} else if isShortenedStackOverflow {
		c.Redirect(302, strings.TrimPrefix(soLink, "https://stackoverflow.com/a/"))
		return
	}

	// if stack exchange, extract the subdomain
	sub := stackExchangeRegex.FindStringSubmatch(soLink)[1]

	c.Redirect(302, fmt.Sprintf("/exchange/%s/%s", sub, strings.TrimPrefix(soLink, fmt.Sprintf("https://%s.stackexchange.com", sub))))
}
