package routes

import (
	"anonymousoverflow/src/types"
	"anonymousoverflow/src/utils"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

func GetImage(c *gin.Context) {

	authorization := c.Query("auth")
	if authorization == "" {
		c.String(400, "Missing auth token")
		return
	}

	// validate the auth token
	token, err := jwt.ParseWithClaims(authorization, &types.ImageProxyClaims{}, func(token *jwt.Token) (interface{}, error) {

		// validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(os.Getenv("JWT_SIGNING_SECRET")), nil
	})
	if err != nil {
		c.String(400, err.Error())
		return
	}

	claims, ok := token.Claims.(*types.ImageProxyClaims)
	if !ok || !token.Valid {
		c.String(400, "Invalid token")
		return
	}

	if claims.Action != "imageProxy" {
		c.String(400, "Invalid action")
		return
	}

	if claims.Exp < time.Now().Unix() {
		c.String(400, "Token expired")
		return
	}

	// download the image
	resp, err := utils.GetWithClearance(claims.ImageURL)
	if err != nil {
		c.AbortWithStatus(500)
		return
	}
	if resp.StatusCode() != http.StatusOK {
		c.AbortWithStatus(http.StatusBadGateway)
		return
	}

	// set the content type
	c.Header("Content-Type", resp.Header().Get("Content-Type"))

	// write the image to the response
	c.Writer.Write(resp.Body())
}
