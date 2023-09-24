package routes

import (
	"anonymousoverflow/config"
	"fmt"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func ChangeOptions(c *gin.Context) {
	name := c.Param("name")

	switch name {
	case "images":
		text := "disabled"
		if c.MustGet("disable_images").(bool) {
			text = "enabled"
		}
		c.SetCookie("disable_images", fmt.Sprintf("%t", !c.MustGet("disable_images").(bool)), 60*60*24*365*10, "/", "", false, true)
		c.HTML(200, "home.html", gin.H{
			"successMessage": "Images are now " + text,
			"theme":          c.MustGet("theme").(string),
			"version":        config.Version,
		})

	case "theme":
		text := "dark"
		if c.MustGet("theme").(string) == "dark" {
			text = "light"
		}

		c.SetCookie("theme", text, 60*60*24*365*10, "/", "", false, true)
		// get redirect url from query
		redirectUrl := c.Query("redirect_url")

		if !strings.HasPrefix(redirectUrl, os.Getenv("APP_URL")) {
			redirectUrl = os.Getenv("APP_URL")
		}

		c.Redirect(302, redirectUrl)

	default:
		c.String(400, "400 Bad Request")
	}
}
