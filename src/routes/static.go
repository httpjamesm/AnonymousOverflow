package routes

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

func StaticContent(c *gin.Context) {
	cleanFilePath := strings.ReplaceAll(c.Param("filepath"), "..", "")

	c.File(fmt.Sprintf("./public/%s", cleanFilePath))
}
