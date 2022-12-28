package routes

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetHome(c *gin.Context) {
	c.HTML(200, "home.html", gin.H{})
}

type urlConversionRequest struct {
	URL string `form:"url" binding:"required"`
}

func PostHome(c *gin.Context) {
	body := urlConversionRequest{}

	if err := c.ShouldBind(&body); err != nil {
		c.HTML(400, "home.html", gin.H{
			"errorMessage": "Invalid request body",
		})
		return
	}

	soLink := body.URL

	// validate URL
	if !strings.HasPrefix(soLink, "https://stackoverflow.com/questions/") {
		c.HTML(400, "home.html", gin.H{
			"errorMessage": "Invalid stack overflow URL",
		})
		return
	}

	// redirect to the proxied thread
	c.Redirect(302, fmt.Sprintf("/questions/%s", strings.TrimPrefix(soLink, "https://stackoverflow.com/questions/")))

}
