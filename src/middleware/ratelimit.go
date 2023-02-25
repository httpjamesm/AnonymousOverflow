package middleware

import (
	"anonymousoverflow/config"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var ipMap = sync.Map{}

func Ratelimit() gin.HandlerFunc {
	return func(c *gin.Context) {

		if os.Getenv("DISABLE_RATELIMIT") == "true" {
			c.Next()
			return
		}

		if strings.HasPrefix(c.Request.URL.Path, "/static") {
			c.Next()
			return
		}

		ip := c.ClientIP()

		// log request count as the value, ip as key
		// if the ip is not in the map, create a new entry with a value of 1
		// if they exceed 30 requests in 1 minute, return a 429

		// get the value from the map
		val, ok := ipMap.Load(ip)
		if !ok {
			// if the ip is not in the map, create a new entry with a value of 1
			ipMap.Store(ip, 1)
			c.Next()
			return
		}

		// if the ip is in the map, increment the value
		ipMap.Store(ip, val.(int)+1)

		// if they exceed 30 requests in 1 minute, return a 429
		if val.(int) > 30 {
			c.HTML(429, "home.html", gin.H{
				"errorMessage": "You have exceeded the request limit. Please try again in a minute.",
				"theme":        c.MustGet("theme").(string),
				"version":      config.Version,
			})
			c.Abort()
			return
		}

		// subtract 1 from the value after 1 minute if the value exists and is greater than 0
		time.AfterFunc(time.Minute, func() {
			val, ok := ipMap.Load(ip)
			if ok && val.(int) > 0 {
				ipMap.Store(ip, val.(int)-1)
			}

			// if the value is 0, delete the entry from the map
			if val.(int) == 0 {
				ipMap.Delete(ip)
			}
		})
	}
}
