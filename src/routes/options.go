package routes

import (
	"fmt"

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
	default:
		c.String(400, "400 Bad Request")
	}

	c.Redirect(302, "/")
}
