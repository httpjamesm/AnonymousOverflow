package routes

import (
	"anonymousoverflow/config"
	"github.com/gin-gonic/gin"
)

func GetVersion(c *gin.Context) {
	c.String(200, config.Version)
}
