package routes

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
)

func RedirectShortenedOverflowURL(c *gin.Context) {
	id := c.Param("id")

	// fetch the stack overflow URL
	client := resty.New()
	client.SetRedirectPolicy(
		resty.RedirectPolicyFunc(func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}),
	)

	resp, err := client.R().Get(fmt.Sprintf("https://www.stackoverflow.com/a/%s", id))
	if err != nil {
		c.HTML(400, "home.html", gin.H{
			"errorMessage": "Unable to fetch stack overflow URL",
			"theme":        c.MustGet("theme").(string),
		})
		return
	}

	if resp.StatusCode() != 302 {
		c.HTML(400, "home.html", gin.H{
			"errorMessage": fmt.Sprintf("Unexpected HTTP status from origin: %d", resp.StatusCode()),
			"theme":        c.MustGet("theme").(string),
		})
		return
	}

	// get the redirect URL
	location := resp.Header().Get("Location")

	c.Redirect(302, fmt.Sprintf("%s%s", os.Getenv("APP_URL"), location))
}
