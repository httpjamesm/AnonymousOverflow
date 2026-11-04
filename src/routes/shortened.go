package routes

import (
	"anonymousoverflow/src/utils"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
)

func RedirectShortenedOverflowURL(c *gin.Context) {
	id := c.Param("id")
	answerId := c.Param("answerId")
	sub := c.Param("sub")

	// fetch the stack overflow URL
	client := resty.New()
	client.SetRedirectPolicy(
		resty.RedirectPolicyFunc(func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}),
	)

	domain := "www.stackoverflow.com"
	if sub != "" {
		host, ok := utils.ExchangeHost(sub)
		if !ok {
			c.HTML(400, "home.html", gin.H{
				"errorMessage": "Invalid Stack Exchange site",
			})
			return
		}
		domain = host
	}
	resp, err := client.R().Get(fmt.Sprintf("https://%s/a/%s/%s", domain, id, answerId))
	if err != nil {
		c.HTML(400, "home.html", gin.H{
			"errorMessage": "Unable to fetch stack overflow URL",
		})
		return
	}

	if resp.StatusCode() != 302 {
		c.HTML(400, "home.html", gin.H{
			"errorMessage": fmt.Sprintf("Unexpected HTTP status from origin: %d", resp.StatusCode()),
		})
		return
	}

	// get the redirect URL
	location := resp.Header().Get("Location")

	redirectPrefix := os.Getenv("APP_URL")
	if sub != "" {
		redirectPrefix += fmt.Sprintf("/exchange/%s", sub)
	}

	c.Redirect(302, fmt.Sprintf("%s%s", redirectPrefix, location))
}
