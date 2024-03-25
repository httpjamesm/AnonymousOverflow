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

var coreRegex = regexp.MustCompile(`(?:https?://)?(?:www\.)?([^/]+)(/(?:questions|q|a)/.+)`)

// Will return `nil` if `rawUrl` is invalid.
func translateUrl(rawUrl string) string {
	coreMatches := coreRegex.FindStringSubmatch(rawUrl)
	if coreMatches == nil {
		return ""
	}

	domain := coreMatches[1]
	rest := coreMatches[2]

	exchange := ""
	if domain == "stackoverflow.com" {
		// No exchange parameter needed.
	} else if sub, found := strings.CutSuffix(domain, ".stackexchange.com"); found {
		if sub == "" {
			return ""
		} else if strings.Contains(sub, ".") {
			// Anything containing dots is interpreted as a full domain, so we use the correct full domain.
			exchange = domain
		} else {
			exchange = sub
		}
	} else {
		exchange = domain
	}

	// Ensure we properly format the return string to avoid double slashes
	if exchange == "" {
		return rest
	} else {
		return fmt.Sprintf("/exchange/%s%s", exchange, rest)
	}
}

func PostHome(c *gin.Context) {
	body := urlConversionRequest{}

	if err := c.ShouldBind(&body); err != nil {
		c.HTML(400, "home.html", gin.H{
			"errorMessage": "Invalid request body",
			"theme":        c.MustGet("theme").(string),
		})
		return
	}

	translated := translateUrl(body.URL)

	if translated == "" {
		c.HTML(400, "home.html", gin.H{
			"errorMessage": "Invalid stack overflow/exchange URL",
			"theme":        c.MustGet("theme").(string),
		})
		return
	}

	c.Redirect(302, translated)
}
