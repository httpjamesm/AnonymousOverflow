package routes

import (
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
		c.String(200, "Images are now %s", text)
	case "theme":
		text := "dark"
		if c.MustGet("theme").(string) == "dark" {
			text = "light"
		}
		c.SetCookie("theme", text, 60*60*24*365*10, "/", "", false, true)
		// get redirect url from query
		redirectUrl := c.Query("redirect_url")
		if redirectUrl == "" {
			redirectUrl = "/"
		}

		if !strings.HasPrefix(redirectUrl, os.Getenv("APP_URL")) {
			redirectUrl = "/"
		}

		c.Redirect(302, redirectUrl)
	default:
		c.String(400, "400 Bad Request")
	}

	c.Redirect(302, "/")
}
